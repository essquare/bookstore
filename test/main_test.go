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
	"bytes"
	"database/sql"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strconv"
	"strings"
	"testing"

	"bookstore/api"
	"bookstore/database"
	"bookstore/model"
	"bookstore/storage"

	"github.com/gorilla/mux"
)

var store *storage.Storage
var db *sql.DB
var r *mux.Router

const (
	contentJSON         = "application/json"
	contentXML          = "application/xml"
	contentAlternateXML = "text/xml"
)

type requestStruct struct {
	user        map[string]interface{}
	url         string
	method      string
	payload     map[string]interface{}
	contentType string
	accept      string
	xmlRoot     string
}

func NewRequest(user map[string]interface{}, url string, method string, payload map[string]interface{}, xmlRoot string, contentType string, accept string) *requestStruct {
	return &requestStruct{user: user, url: url, method: method, payload: payload, xmlRoot: xmlRoot, contentType: contentType, accept: accept}
}
func (r *requestStruct) createXMLString() string {
	str := fmt.Sprintf("<%s>", r.xmlRoot)
	for key, value := range r.payload {
		str += fmt.Sprintf("<%s>%v</%s>", key, value, key)
	}
	str += fmt.Sprintf("</%s>", r.xmlRoot)
	return str
}

func (r *requestStruct) unmarshal(t *testing.T, response *httptest.ResponseRecorder, m interface{}) {
	if len(response.Body.Bytes()) != 0 {
		switch r.contentType {
		case contentJSON:
			if o, ok := m.(api.ListContainer); ok {
				list := o.InternalList()
				err := json.Unmarshal(response.Body.Bytes(), &list)
				if err != nil {
					t.Fatalf("Problem unmarshaling request: %v\n", err)
				}
			} else {
				err := json.Unmarshal(response.Body.Bytes(), &m)
				if err != nil {
					t.Fatalf("Problem unmarshaling request: %v\n", err)
				}
			}
		case contentXML, contentAlternateXML:
			err := xml.Unmarshal(response.Body.Bytes(), &m)
			if err != nil {
				t.Fatalf("Problem unmarshaling request: %v\n", err)
			}
		default:
			if o, ok := m.(api.ListContainer); ok {
				list := o.InternalList()
				err := json.Unmarshal(response.Body.Bytes(), &list)
				if err != nil {
					t.Fatalf("Problem unmarshaling request: %v\n", err)
				}
			} else {
				err := json.Unmarshal(response.Body.Bytes(), &m)
				if err != nil {
					t.Fatalf("Problem unmarshaling request: %v\n", err)
				}
			}
		}
	}
	m = nil
}
func (r *requestStruct) makeRequest(t *testing.T) *httptest.ResponseRecorder {
	token := getUserJWT(t, r.user)

	var body io.Reader
	if r.payload != nil {
		switch r.contentType {
		case contentJSON:
			str, err := json.Marshal(r.payload)
			if err != nil {
				t.Fatalf("Problem marshaling request: %v\n", err)
			}
			body = bytes.NewBuffer([]byte(str))
		case contentXML, contentAlternateXML:
			str := r.createXMLString()
			body = bytes.NewBuffer([]byte(str))
		default:
			// in order to test some edge cases just default to json
			str, err := json.Marshal(r.payload)
			if err != nil {
				t.Fatalf("Problem marshaling request: %v\n", err)
			}
			body = bytes.NewBuffer([]byte(str))
		}
	}
	request, err := http.NewRequest(r.method, r.url, body)
	if err != nil {
		t.Fatalf("Problem creating request: %v\n", err)
	}
	request.Header.Set("Content-Type", r.contentType)
	request.Header.Set("Accept", r.accept)
	request = addBearerToken(request, token)
	return executeRequest(request)
}

func TestMain(m *testing.M) {
	tmpfile, err := ioutil.TempFile("", "testdatabase.*.sqlite")
	if err != nil {
		log.Fatal(err)
	}

	db, err = database.NewDatabaseConnection(tmpfile.Name())
	if err != nil {
		log.Fatalf("Unable to initialize database connection pool: %v", err)
	}

	store = storage.NewStorage(db)
	if err = store.Ping(); err != nil {
		log.Fatalf("Unable to connect to the database: %v", err)
	}

	if err = database.Migrate(db); err != nil {
		log.Fatalf(`%v`, err)
	}

	if err = database.CurrentDBSchema(db); err != nil {
		log.Fatalf("The DB version should have been correct, %v", err)
	}

	r = mux.NewRouter()
	api.Serve(r, store)
	code := m.Run()

	// os.Exit() does not respect defer statements
	db.Close()
	os.Remove(tmpfile.Name()) // clean up
	os.Exit(code)
}

func resetDatabase(t *testing.T) {
	_, err := db.Exec("DELETE FROM books")
	if err != nil {
		t.Fatalf("Problem cleaning the database: %v\n", err)
	}
	_, err = db.Exec("DELETE FROM sqlite_sequence WHERE `name` = 'books'")
	if err != nil {
		t.Fatalf("Problem cleaning the database: %v\n", err)
	}
	_, err = db.Exec("DELETE FROM users")
	if err != nil {
		t.Fatalf("Problem cleaning the database: %v\n", err)
	}
	_, err = db.Exec("DELETE FROM sqlite_sequence WHERE `name` = 'users'")
	if err != nil {
		t.Fatalf("Problem cleaning the database: %v\n", err)
	}
}

func createDefaultAdmin(t *testing.T) map[string]interface{} {
	user, err := store.CreateUser(&model.UserCreationRequest{
		Username:  "admin",
		Password:  "test123",
		Pseudonym: "Admin User",
		IsAdmin:   true,
	})
	if err != nil {
		t.Fatalf("Problem creating the admin user: %v\n", err)
	}

	var m map[string]interface{}
	jsonUser, err := json.Marshal(user)
	if err != nil {
		t.Fatalf("Problem creating the admin user: %v\n", err)
	}
	json.Unmarshal(jsonUser, &m)
	m["password"] = "test123"
	m["id"] = int64(m["id"].(float64))
	return m
}

func createSimpleUser(t *testing.T, username, pseudonym string) map[string]interface{} {
	user, err := store.CreateUser(&model.UserCreationRequest{
		Username:  username,
		Password:  "test123",
		Pseudonym: pseudonym,
		IsAdmin:   true,
	})
	if err != nil {
		t.Fatalf("Problem creating user: %v\n", err)
	}

	var m map[string]interface{}
	jsonUser, err := json.Marshal(user)
	if err != nil {
		t.Fatalf("Problem creating user: %v\n", err)
	}
	json.Unmarshal(jsonUser, &m)
	m["password"] = "test123"
	m["id"] = int64(m["id"].(float64))
	return m
}

func getUserJWT(t *testing.T, user map[string]interface{}) string {
	data := url.Values{}
	data.Set("username", user["username"].(string))
	data.Set("password", user["password"].(string))

	request, err := http.NewRequest(http.MethodPost, "/authenticate", strings.NewReader(data.Encode())) // URL-encoded payload
	if err != nil {
		t.Fatalf("Problem getting jwt token for admin: %v\n", err)
	}
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	request.Header.Add("Content-Length", strconv.Itoa(len(data.Encode())))
	response := executeRequest(request)
	if response.Code != http.StatusOK {
		t.Fatalf("Expected 200: %d received", response.Code)
	}
	var m map[string]string
	json.Unmarshal(response.Body.Bytes(), &m)
	if m["token"] == "" {
		t.Fatalf("Expected token: empty string received")
	}

	return m["token"]
}

func addBearerToken(request *http.Request, token string) *http.Request {
	request.Header.Add("Authorization", "Bearer "+token)
	return request
}
func executeRequest(req *http.Request) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	return rr
}

func checkResponseCode(t *testing.T, actual, expected int) {
	if expected != actual {
		t.Fatalf("Expected response code %d. Got %d\n", expected, actual)
	}
}

func checkUser(t *testing.T, user map[string]interface{}, userResponse *model.User) {
	if userResponse.Username != user["username"].(string) {
		t.Fatalf("Expected username %s. Got %s\n", user["username"].(string), userResponse.Username)
	}
	if userResponse.Pseudonym != user["pseudonym"].(string) {
		t.Fatalf("Expected pseudonym %s. Got %s\n", user["pseudonym"].(string), userResponse.Pseudonym)
	}

	if userResponse.IsAdmin != user["is_admin"].(bool) {
		t.Fatalf("Expected is_admin %t. Got %t\n", user["is_admin"].(bool), userResponse.IsAdmin)
	}

	if userResponse.ID != user["id"].(int64) {
		t.Fatalf("Expected id %d. Got %d\n", user["id"].(int64), userResponse.ID)
	}
}

func checkBooks(t *testing.T, expectedBooks []map[string]interface{}, booksResponse *model.Books) {
	if len(expectedBooks) != len(booksResponse.Books) {
		t.Fatalf("Expected count of books %d. Got %d\n", len(expectedBooks), len(booksResponse.Books))
	}

	for idx, expectedBooks := range expectedBooks {
		checkBook(t, expectedBooks, &booksResponse.Books[idx])
	}
}

func checkBook(t *testing.T, book map[string]interface{}, bookResponse *model.Book) {
	if bookResponse.Price != book["price"].(int64) {
		t.Fatalf("Expected price %d. Got %d\n", book["price"].(int64), bookResponse.Price)
	}
	if bookResponse.Title != book["title"].(string) {
		t.Fatalf("Expected title %s. Got %s\n", book["title"].(string), bookResponse.Title)
	}

	if bookResponse.Description != book["description"].(string) {
		t.Fatalf("Expected description %s. Got %s\n", book["description"].(string), bookResponse.Description)
	}

	if bookResponse.ImageURL != book["image_url"].(string) {
		t.Fatalf("Expected image_url %s. Got %s\n", book["image_url"].(string), bookResponse.ImageURL)
	}

	if bookResponse.UserID != book["user_id"].(int64) {
		t.Fatalf("Expected user_id %d. Got %d\n", book["user_id"].(int64), bookResponse.UserID)
	}

	if bookResponse.ID != book["id"].(int64) {
		t.Fatalf("Expected id %d. Got %d\n", book["id"].(int64), bookResponse.ID)
	}
}

func getXMLError(errMsg string) string {
	return "<error><error_message>" + errMsg + "</error_message></error>"
}

func getJSONError(errMsg string) string {
	return `{"error_message":"` + errMsg + `"}`
}
