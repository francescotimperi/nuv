# Licensed to the Apache Software Foundation (ASF) under one
# or more contributor license agreements.  See the NOTICE file
# distributed with this work for additional information
# regarding copyright ownership.  The ASF licenses this file
# to you under the Apache License, Version 2.0 (the
# "License"); you may not use this file except in compliance
# with the License.  You may obtain a copy of the License at
#
#   http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing,
# software distributed under the License is distributed on an
# "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
# KIND, either express or implied.  See the License for the
# specific language governing permissions and limitations
# under the License.

setup() {
    load 'test_helper/bats-support/load'
    load 'test_helper/bats-assert/load'
    export NO_COLOR=1
}

@test "-js -h" {
    run nuv -js -h
    assert_line "Usage: nuv -js FILE.js"
    assert_line "Interpret and run Javascript code."
}

@test "-js with simple file.js" {
    run nuv -js testdata/js_simple_test.js
    assert_line "2 + 2 = 4"
    assert_success
}

@test "-js with js function" {
    run nuv -js testdata/js_function_test.js
    assert_line "3"
    assert_success
}
