// Copyright 2021 essquare GmbH
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package test

import (
	"fmt"
	"net/http"
	"testing"

	"bookstore.app/model"
)

func createUser(t *testing.T, caller map[string]interface{}, user *map[string]interface{}, contentType string) {
	var m model.User
	r := NewRequest(caller, fmt.Sprintf("/users"), http.MethodPost, *user, "User", contentType, contentType)
	response := r.makeRequest(t)

	checkResponseCode(t, response.Code, http.StatusCreated)

	r.unmarshal(t, response, &m)
	(*user)["id"] = m.ID
	checkUser(t, *user, &m)

	getUser(t, caller, user, contentType)

}

func updateUser(t *testing.T, caller map[string]interface{}, user *map[string]interface{}, change map[string]interface{}, contentType string) {
	var m model.User
	r := NewRequest(caller, fmt.Sprintf("/users/%v", (*user)["id"]), http.MethodPut, change, "User", contentType, contentType)
	response := r.makeRequest(t)
	checkResponseCode(t, response.Code, http.StatusOK)

	r.unmarshal(t, response, &m)

	for key, value := range change {
		(*user)[key] = value
	}

	checkUser(t, *user, &m)

	getUser(t, caller, user, contentType)

}

func deleteUser(t *testing.T, caller map[string]interface{}, user *map[string]interface{}, contentType string) {
	r := NewRequest(caller, fmt.Sprintf("/users/%v", (*user)["id"]), http.MethodDelete, nil, "User", contentType, contentType)
	response := r.makeRequest(t)
	checkResponseCode(t, response.Code, http.StatusNoContent)

	getNonExistentUser(t, caller, user, contentType)
}

func getNonExistentUser(t *testing.T, caller map[string]interface{}, user *map[string]interface{}, contentType string) {
	r := NewRequest(caller, fmt.Sprintf("/users/%v", (*user)["id"]), http.MethodGet, nil, "User", contentType, contentType)
	response := r.makeRequest(t)
	checkResponseCode(t, response.Code, http.StatusNotFound)
}

func getUser(t *testing.T, caller map[string]interface{}, user *map[string]interface{}, contentType string) {
	var m model.User

	r := NewRequest(caller, fmt.Sprintf("/users/%v", (*user)["id"]), http.MethodGet, nil, "User", contentType, contentType)
	response := r.makeRequest(t)
	checkResponseCode(t, response.Code, http.StatusOK)

	r.unmarshal(t, response, &m)

	checkUser(t, *user, &m)
}
func TestGeneralUserOperations(t *testing.T) {
	resetDatabase(t)
	admin := createDefaultAdmin(t)

	contentTypes := []string{contentXML, contentJSON, contentAlternateXML}

	for _, contentType := range contentTypes {

		user := map[string]interface{}{
			"username":  "testuser123",
			"pseudonym": "Jack London",
			"password":  "test123",
			"is_admin":  false,
		}

		createUser(t, admin, &user, contentType)

		change := map[string]interface{}{
			"pseudonym": "Jack Kerouac",
		}

		updateUser(t, admin, &user, change, contentType)

		deleteUser(t, admin, &user, contentType)

		createUser(t, admin, &user, contentType)

		deleteUser(t, admin, &user, contentType)
	}
}

func TestGetAdminUser(t *testing.T) {
	resetDatabase(t)
	admin := createDefaultAdmin(t)

	contentTypes := []string{contentXML, contentJSON, contentAlternateXML}

	for _, contentType := range contentTypes {
		getUser(t, admin, &admin, contentType)
	}
}
