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

	"bookstore/model"
)

func createUser(t *testing.T, caller map[string]interface{}, user *map[string]interface{}, contentType string) {
	var m model.User
	r := NewRequest(caller, "/users", http.MethodPost, *user, "user", contentType, contentType)
	response := r.makeRequest(t)

	checkResponseCode(t, response.Code, http.StatusCreated)

	r.unmarshal(t, response, &m)
	(*user)["id"] = m.ID
	checkUser(t, *user, &m)

	getUser(t, caller, user, contentType)
}

func createUserWithError(t *testing.T, caller map[string]interface{}, user *map[string]interface{}, contentType string, errorCode int64, errorString string) {
	r := NewRequest(caller, "/users", http.MethodPost, *user, "user", contentType, contentType)
	response := r.makeRequest(t)

	checkResponseCode(t, response.Code, int(errorCode))

	var expectedString string
	switch contentType {
	case contentJSON:
		expectedString = getJSONError(errorString)
	case contentXML, contentAlternateXML:
		expectedString = getXMLError(errorString)
	default:
		expectedString = getJSONError(errorString)
	}
	if response.Body.String() != expectedString {
		t.Fatalf("Error does not match. Expected %s - received %s\n", expectedString, response.Body.String())
	}
}
func updateUser(t *testing.T, caller map[string]interface{}, user *map[string]interface{}, change map[string]interface{}, contentType string) {
	var m model.User
	r := NewRequest(caller, fmt.Sprintf("/users/%v", (*user)["id"]), http.MethodPut, change, "user", contentType, contentType)
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
	r := NewRequest(caller, fmt.Sprintf("/users/%v", (*user)["id"]), http.MethodDelete, nil, "user", contentType, contentType)
	response := r.makeRequest(t)
	checkResponseCode(t, response.Code, http.StatusNoContent)

	getNonExistentUser(t, caller, user, contentType)
}

func getNonExistentUser(t *testing.T, caller map[string]interface{}, user *map[string]interface{}, contentType string) {
	r := NewRequest(caller, fmt.Sprintf("/users/%v", (*user)["id"]), http.MethodGet, nil, "user", contentType, contentType)
	response := r.makeRequest(t)
	checkResponseCode(t, response.Code, http.StatusNotFound)
}

func getUser(t *testing.T, caller map[string]interface{}, user *map[string]interface{}, contentType string) {
	var m model.User

	r := NewRequest(caller, fmt.Sprintf("/users/%v", (*user)["id"]), http.MethodGet, nil, "user", contentType, contentType)
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

func TestGeneralErrorCases(t *testing.T) {
	resetDatabase(t)
	admin := createDefaultAdmin(t)

	user := map[string]interface{}{
		"username":  "testuser123",
		"pseudonym": "Jack London",
		"password":  "test123",
		"is_admin":  false,
	}

	createUser(t, admin, &user, contentJSON)
	contentTypes := []string{contentXML, contentJSON, contentAlternateXML}

	for _, contentType := range contentTypes {

		duplicateUsernameUser := map[string]interface{}{
			"username":  "testuser123",
			"pseudonym": "Mark Twain",
			"password":  "test123",
			"is_admin":  false,
		}

		createUserWithError(t, admin, &duplicateUsernameUser, contentType, http.StatusBadRequest, "user_already_exists")

		emptyUsernameUser := map[string]interface{}{
			"username":  "",
			"pseudonym": "Mark Twain",
			"password":  "test123",
			"is_admin":  false,
		}

		createUserWithError(t, admin, &emptyUsernameUser, contentType, http.StatusBadRequest, "user_mandatory_fields:username")

		duplicatePseudonymUser := map[string]interface{}{
			"username":  "testuser344",
			"pseudonym": "Jack London",
			"password":  "test123",
			"is_admin":  false,
		}

		createUserWithError(t, admin, &duplicatePseudonymUser, contentType, http.StatusBadRequest, "user_already_exists")

		emptyPseudonymUser := map[string]interface{}{
			"username":  "testuser344",
			"pseudonym": "",
			"password":  "test123",
			"is_admin":  false,
		}

		createUserWithError(t, admin, &emptyPseudonymUser, contentType, http.StatusBadRequest, "user_mandatory_fields:pseudonym")

		shortPasswordUser := map[string]interface{}{
			"username":  "testuser344",
			"pseudonym": "Mark Twain",
			"password":  "test",
			"is_admin":  false,
		}

		createUserWithError(t, admin, &shortPasswordUser, contentType, http.StatusBadRequest, "password_min_length")
	}

}
