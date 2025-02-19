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
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"

	"github.com/zalando/go-keyring"
)

type LoginResult struct {
	Login   string
	Auth    string
	ApiHost string
}

const usage = `Usage:
nuv login <apihost> [<user>]

Login to a Nuvolaris instance. If no user is specified, the default user "nuvolaris" is used.

Options:
  -h, --help   Show usage`

const whiskLoginPath = "/api/v1/web/whisk-system/nuv/login"
const defaultUser = "nuvolaris"
const nuvSecretServiceName = "nuvolaris"

func LoginCmd() (*LoginResult, error) {
	flag.Usage = func() {
		fmt.Println(usage)
	}

	var helpFlag bool
	flag.BoolVar(&helpFlag, "h", false, "Show usage")
	flag.BoolVar(&helpFlag, "help", false, "Show usage")
	flag.Parse()

	if helpFlag {
		flag.Usage()
		return nil, nil
	}

	args := flag.Args()

	if len(args) == 0 {
		flag.Usage()
		return nil, errors.New("missing apihost")
	}

	fmt.Print("Enter Password: ")
	pwd, err := AskPassword()
	if err != nil {
		return nil, err
	}
	fmt.Println()
	apihost := args[0]
	url := apihost + whiskLoginPath
	user := defaultUser
	if len(args) > 1 {
		user = args[1]
	}

	creds, err := doLogin(url, user, pwd)
	if err != nil {
		return nil, err
	}

	if err := storeCredentials(creds); err != nil {
		return nil, err
	}

	auth, err := keyring.Get(nuvSecretServiceName, "AUTH")
	if err != nil {
		return nil, err
	}

	return &LoginResult{
		Login:   user,
		Auth:    auth,
		ApiHost: apihost,
	}, nil
}

func doLogin(url, user, password string) (map[string]string, error) {
	data := map[string]string{
		"login":    user,
		"password": password,
	}
	loginJson, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(loginJson))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("login failed with status code %d", resp.StatusCode)
		}
		return nil, fmt.Errorf("login failed (%d): %s", resp.StatusCode, string(body))
	}

	var creds map[string]string
	err = json.NewDecoder(resp.Body).Decode(&creds)
	if err != nil {
		return nil, errors.New("failed to decode response from login request")
	}

	return creds, nil
}

func storeCredentials(creds map[string]string) error {
	for k, v := range creds {
		err := keyring.Set(nuvSecretServiceName, k, v)
		if err != nil {
			return err
		}
	}

	return nil
}
