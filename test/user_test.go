// Copyright 2021 gergan
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
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"log"
	"net/http"
	"testing"

	"bookstore.app/model"
)

func TestGetAdminUser(t *testing.T) {
	err := resetDatabase()
	if err != nil {
		log.Fatalf("Problem cleaning the database: %v\n", err)
	}

	admin, err := createDefaultAdmin()
	if err != nil {
		log.Fatalf("Problem creating the admin user: %v\n", err)
	}
	token, err := getUserJWT(admin)
	if err != nil {
		t.Errorf("Problem getting jwt token for admin: %v\n", err)
	}

	request, err := http.NewRequest(http.MethodGet, fmt.Sprintf("/users/%d", admin.ID), nil)
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Accept", "application/json")
	if err != nil {
		t.Errorf("Problem creating request: %v\n", err)
	}
	request = addBearerToken(request, token)
	response := executeRequest(request)
	checkResponseCode(t, response.Code, http.StatusOK)

	var m model.User
	json.Unmarshal(response.Body.Bytes(), &m)

	checkUser(t, admin, &m)

}

func TestGetAdminUserXML(t *testing.T) {
	err := resetDatabase()
	if err != nil {
		log.Fatalf("Problem cleaning the database: %v\n", err)
	}

	admin, err := createDefaultAdmin()
	if err != nil {
		log.Fatalf("Problem creating the admin user: %v\n", err)
	}
	token, err := getUserJWT(admin)
	if err != nil {
		t.Errorf("Problem getting jwt token for admin: %v\n", err)
	}

	request, err := http.NewRequest(http.MethodGet, fmt.Sprintf("/users/%d", admin.ID), nil)
	request.Header.Set("Content-Type", "application/xml")
	request.Header.Set("Accept", "application/xml")
	if err != nil {
		t.Errorf("Problem creating request: %v\n", err)
	}
	request = addBearerToken(request, token)
	response := executeRequest(request)
	checkResponseCode(t, response.Code, http.StatusOK)

	var m model.User
	xml.Unmarshal(response.Body.Bytes(), &m)

	checkUser(t, admin, &m)

}
func TestUserOperations(t *testing.T) {
	err := resetDatabase()
	if err != nil {
		log.Fatalf("Problem cleaning the database: %v\n", err)
	}

	admin, err := createDefaultAdmin()
	if err != nil {
		log.Fatalf("Problem creating the admin user: %v\n", err)
	}
	token, err := getUserJWT(admin)
	if err != nil {
		t.Errorf("Problem getting jwt token for admin: %v\n", err)
	}

	user := &model.User{
		Username:  "testuser123",
		Pseudonym: "Jack London",
		Password:  "test123",
		IsAdmin:   false,
	}
	jsonStr := createJSONString(user)
	request, err := http.NewRequest(http.MethodPost, "/users", bytes.NewBuffer([]byte(jsonStr)))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Accept", "application/json")
	if err != nil {
		t.Errorf("Problem creating request: %v\n", err)
	}
	request = addBearerToken(request, token)
	response := executeRequest(request)
	checkResponseCode(t, response.Code, http.StatusCreated)

	var m model.User
	json.Unmarshal(response.Body.Bytes(), &m)
	user.ID = m.ID
	checkUser(t, user, &m)

	request, err = http.NewRequest(http.MethodPut, fmt.Sprintf("/users/%d", user.ID), bytes.NewBuffer([]byte(`{"pseudonym":"Jack Kerouac"}`)))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Accept", "application/json")
	if err != nil {
		t.Errorf("Problem creating request: %v\n", err)
	}
	request = addBearerToken(request, token)
	response = executeRequest(request)
	checkResponseCode(t, response.Code, http.StatusOK)

	json.Unmarshal(response.Body.Bytes(), &m)
	user.Pseudonym = "Jack Kerouac"
	checkUser(t, user, &m)

	request, err = http.NewRequest(http.MethodGet, fmt.Sprintf("/users/%d", user.ID), nil)
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Accept", "application/json")
	if err != nil {
		t.Errorf("Problem creating request: %v\n", err)
	}
	request = addBearerToken(request, token)
	response = executeRequest(request)
	checkResponseCode(t, response.Code, http.StatusOK)

	json.Unmarshal(response.Body.Bytes(), &m)

	checkUser(t, user, &m)

	request, err = http.NewRequest(http.MethodDelete, fmt.Sprintf("/users/%d", user.ID), nil)
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Accept", "application/json")
	if err != nil {
		t.Errorf("Problem creating request: %v\n", err)
	}
	request = addBearerToken(request, token)
	response = executeRequest(request)
	checkResponseCode(t, response.Code, http.StatusNoContent)

	request, err = http.NewRequest(http.MethodGet, fmt.Sprintf("/users/%d", user.ID), nil)
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Accept", "application/json")
	if err != nil {
		t.Errorf("Problem creating request: %v\n", err)
	}
	request = addBearerToken(request, token)
	response = executeRequest(request)
	checkResponseCode(t, response.Code, http.StatusNotFound)

}

func TestUserOperationsXML(t *testing.T) {
	err := resetDatabase()
	if err != nil {
		log.Fatalf("Problem cleaning the database: %v\n", err)
	}

	admin, err := createDefaultAdmin()
	if err != nil {
		log.Fatalf("Problem creating the admin user: %v\n", err)
	}
	token, err := getUserJWT(admin)
	if err != nil {
		t.Errorf("Problem getting jwt token for admin: %v\n", err)
	}

	user := &model.User{
		Username:  "testuser123",
		Pseudonym: "Jack London",
		Password:  "test123",
		IsAdmin:   false,
	}
	xmlStr := createXMLString(user)
	request, err := http.NewRequest(http.MethodPost, "/users", bytes.NewBuffer([]byte(xmlStr)))
	request.Header.Set("Content-Type", "application/xml")
	request.Header.Set("Accept", "application/xml")
	if err != nil {
		t.Errorf("Problem creating request: %v\n", err)
	}
	request = addBearerToken(request, token)
	response := executeRequest(request)
	checkResponseCode(t, response.Code, http.StatusCreated)

	var m model.User
	xml.Unmarshal(response.Body.Bytes(), &m)
	user.ID = m.ID
	checkUser(t, user, &m)

	request, err = http.NewRequest(http.MethodPut, fmt.Sprintf("/users/%d", user.ID), bytes.NewBuffer([]byte(`<User><pseudonym>Jack Kerouac</pseudonym></User>`)))
	request.Header.Set("Content-Type", "application/xml")
	request.Header.Set("Accept", "application/xml")
	if err != nil {
		t.Errorf("Problem creating request: %v\n", err)
	}
	request = addBearerToken(request, token)
	response = executeRequest(request)
	checkResponseCode(t, response.Code, http.StatusOK)

	xml.Unmarshal(response.Body.Bytes(), &m)
	user.Pseudonym = "Jack Kerouac"
	checkUser(t, user, &m)

	request, err = http.NewRequest(http.MethodGet, fmt.Sprintf("/users/%d", user.ID), nil)
	request.Header.Set("Content-Type", "application/xml")
	request.Header.Set("Accept", "application/xml")
	if err != nil {
		t.Errorf("Problem creating request: %v\n", err)
	}
	request = addBearerToken(request, token)
	response = executeRequest(request)
	checkResponseCode(t, response.Code, http.StatusOK)

	xml.Unmarshal(response.Body.Bytes(), &m)

	checkUser(t, user, &m)

	request, err = http.NewRequest(http.MethodDelete, fmt.Sprintf("/users/%d", user.ID), nil)
	request.Header.Set("Content-Type", "application/xml")
	request.Header.Set("Accept", "application/xml")
	if err != nil {
		t.Errorf("Problem creating request: %v\n", err)
	}
	request = addBearerToken(request, token)
	response = executeRequest(request)
	checkResponseCode(t, response.Code, http.StatusNoContent)

	request, err = http.NewRequest(http.MethodGet, fmt.Sprintf("/users/%d", user.ID), nil)
	request.Header.Set("Content-Type", "application/xml")
	request.Header.Set("Accept", "application/xml")
	if err != nil {
		t.Errorf("Problem creating request: %v\n", err)
	}
	request = addBearerToken(request, token)
	response = executeRequest(request)
	checkResponseCode(t, response.Code, http.StatusNotFound)

}
