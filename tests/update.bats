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
    export NUV_NO_LOG_PREFIX=1
}

@test "nuv -update" {
    run nuv -update
    assert_line "Nuvfiles downloaded successfully"
    assert_success
}

@test "nuv -update with old version warns" {
    NUV_VERSION=0.2.0 run nuv -update
    assert_line "Your nuv version 0.2.0 is older than the required version in nuvroot.json."
    assert_line "Please update nuv to the latest version."
    assert_success
}

@test "nuv -update with bad version" {
    NUV_VERSION=notsemver run nuv -update
    assert_line "Unable to validate nuv version notsemver : Invalid Semantic Version"
    assert_success
}

@test "nuv -update with newer version" {
    NUV_VERSION=1.2.3 run nuv -update
    assert_line "Your tasks are already up to date!"
    assert_success
}
