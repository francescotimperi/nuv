// Licensed to the Apache Software Foundation (ASF) under one
// or more contributor license agreements.  See the NOTICE file
// distributed with this work for additional information
// regarding copyright ownership.  The ASF licenses this file
// to you under the Apache License, Version 2.0 (the
// "License"); you may not use this file except in compliance
// with the License.  You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package auth

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/zalando/go-keyring"
)

// setupMockServer sets up a new mock HTTP server with the given test data and expected response
func setupMockServer(t *testing.T, expectedLogin, expectedPass, expectedRes string) *httptest.Server {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var requestBody map[string]string
		err := json.NewDecoder(r.Body).Decode(&requestBody)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if login, ok := requestBody["login"]; !ok {
			t.Error("expected login field in request body")
		} else if login != expectedLogin {
			t.Errorf("expected login %s, got %s", expectedLogin, login)
		}
		if password, ok := requestBody["password"]; !ok {
			t.Error("expected password field in request body")
		} else if password != expectedPass {
			t.Errorf("expected password %s, got %s", expectedPass, password)
		}

		_, _ = w.Write([]byte(expectedRes))
	}))

	return server
}
func ExampleLoginCmd_noArgs() {
	os.Args = []string{"login"}
	_, err := LoginCmd()
	fmt.Println(err)
	// Output:
	// Usage:
	// nuv login <apihost> [<user>]
	// <nil>
}

func TestLoginCmd(t *testing.T) {
	keyring.MockInit()

	t.Run("error: returns error when empty password", func(t *testing.T) {
		oldPwdReader := pwdReader
		pwdReader = &stubPasswordReader{
			Password:    "",
			ReturnError: false,
		}
		os.Args = []string{"login", "fakeApiHost", "fakeUser"}
		res, err := LoginCmd()
		pwdReader = oldPwdReader
		if err == nil {
			t.Error("Expected error, got nil")
		}
		if err.Error() != "password is empty" {
			t.Errorf("Expected error to be 'password is empty', got %s", err.Error())
		}
		if res != nil {
			t.Errorf("Expected response to be nil, got %v", res)
		}
	})

	t.Run("with only apihost adds received credentials to secret store", func(t *testing.T) {
		mockServer := setupMockServer(t, "nuvolaris", "a password", "{\"fakeCred\": \"test\"}")
		defer mockServer.Close()

		oldPwdReader := pwdReader
		pwdReader = &stubPasswordReader{
			Password:    "a password",
			ReturnError: false,
		}

		os.Args = []string{"login", mockServer.URL}
		loginResult, err := LoginCmd()
		pwdReader = oldPwdReader

		if err != nil {
			t.Errorf("Expected no error, got %s", err.Error())
		}

		if loginResult == nil {
			t.Error("Expected result, got nil")
		}

		cred, err := keyring.Get(nuvSecretServiceName, "fakeCred")
		if err != nil {
			t.Errorf("Expected no error, got %s", err.Error())
		}

		if cred != "test" {
			t.Errorf("Expected test, got %s", cred)
		}
	})

	t.Run("with apihost and user adds received credentials to secret store", func(t *testing.T) {
		mockServer := setupMockServer(t, "a user", "a password", "{ \"fakeCred\": \"test\"}")
		defer mockServer.Close()

		oldPwdReader := pwdReader
		pwdReader = &stubPasswordReader{
			Password:    "a password",
			ReturnError: false,
		}

		os.Args = []string{"login", mockServer.URL, "a user"}
		loginResult, err := LoginCmd()
		pwdReader = oldPwdReader

		if err != nil {
			t.Errorf("Expected no error, got %s", err.Error())
		}

		if loginResult == nil {
			t.Error("Expected result, got nil")
		}

		cred, err := keyring.Get(nuvSecretServiceName, "fakeCred")
		if err != nil {
			t.Errorf("Expected no error, got %s", err.Error())
		}

		if cred != "test" {
			t.Errorf("Expected test, got %s", cred)
		}
	})

	t.Run("error when response body is invalid", func(t *testing.T) {
		mockServer := setupMockServer(t, "a user", "a password", "invalid json")
		defer mockServer.Close()

		oldPwdReader := pwdReader
		pwdReader = &stubPasswordReader{
			Password:    "a password",
			ReturnError: false,
		}
		os.Args = []string{"login", mockServer.URL, "a user"}
		loginResult, err := LoginCmd()
		pwdReader = oldPwdReader

		if err == nil {
			t.Error("Expected error, got nil")
		}
		if err.Error() != "failed to decode response from login request" {
			t.Errorf("Expected error to be 'failed to decode response from login request', got %s", err.Error())
		}

		if loginResult != nil {
			t.Errorf("Expected loginResult to be nil, got %v", loginResult)
		}
	})
}

func Test_doLogin(t *testing.T) {
	mockServer := setupMockServer(t, "a user", "a password", "{\"fakeCred\": \"test\"}")
	defer mockServer.Close()

	cred, err := doLogin(mockServer.URL, "a user", "a password")
	if err != nil {
		t.Errorf("Expected no error, got %s", err.Error())
	}

	if cred["fakeCred"] != "test" {
		t.Errorf("Expected test, got %s", cred["fakeCred"])
	}
}
func Test_storeCredentials(t *testing.T) {
	keyring.MockInit()

	fakeCreds := make(map[string]string)
	fakeCreds["AUTH"] = "fakeValue"
	fakeCreds["REDIS_URL"] = "fakeValue"
	fakeCreds["MONGODB"] = "fakeValue"

	err := storeCredentials(fakeCreds)
	if err != nil {
		t.Errorf("Expected no error, got %s", err.Error())
	}

	for k := range fakeCreds {
		cred, err := keyring.Get(nuvSecretServiceName, k)
		if err != nil {
			t.Errorf("Expected no error, got %s", err.Error())
		}
		if cred != fakeCreds[k] {
			t.Errorf("Expected %s, got %s", fakeCreds[k], cred)
		}
	}
}
