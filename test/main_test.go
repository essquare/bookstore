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
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strconv"
	"strings"
	"testing"

	"bookstore.app/api"
	"bookstore.app/database"
	"bookstore.app/model"
	"bookstore.app/storage"
	"github.com/gorilla/mux"
)

var store *storage.Storage
var db *sql.DB
var r *mux.Router

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

func resetDatabase() error {

	_, err := db.Exec("DELETE FROM books")
	if err != nil {
		return err
	}
	_, err = db.Exec("DELETE FROM sqlite_sequence WHERE `name` = 'books'")
	if err != nil {
		return err
	}
	_, err = db.Exec("DELETE FROM users")
	if err != nil {
		return err
	}
	_, err = db.Exec("DELETE FROM sqlite_sequence WHERE `name` = 'users'")
	if err != nil {
		return err
	}
	return nil
}

func createDefaultAdmin() (*model.User, error) {
	user, err := store.CreateUser(&model.UserCreationRequest{
		Username:  "admin",
		Password:  "test123",
		Pseudonym: "Admin User",
		IsAdmin:   true,
	})
	if err != nil {
		return nil, err
	}
	user.Password = "test123"
	return user, nil
}

func getUserJWT(user *model.User) (string, error) {
	data := url.Values{}
	data.Set("username", user.Username)
	data.Set("password", user.Password)

	request, err := http.NewRequest(http.MethodPost, "/authenticate", strings.NewReader(data.Encode())) // URL-encoded payload
	if err != nil {
		return "", err
	}
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	request.Header.Add("Content-Length", strconv.Itoa(len(data.Encode())))
	response := executeRequest(request)
	if response.Code == http.StatusOK {

		var m map[string]string
		json.Unmarshal(response.Body.Bytes(), &m)
		if m["token"] != "" {
			return m["token"], nil
		}
		return "", fmt.Errorf("Expected token: empty string received")
	}
	return "", fmt.Errorf("Expected 200: %d received", response.Code)
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
		t.Errorf("Expected response code %d. Got %d\n", expected, actual)
	}
}

func checkUser(t *testing.T, user *model.User, userResponse *model.User) {
	if userResponse.Username != user.Username {
		t.Errorf("Expected username %s. Got %s\n", user.Username, userResponse.Username)
	}
	if userResponse.Pseudonym != user.Pseudonym {
		t.Errorf("Expected pseudonym %s. Got %s\n", user.Pseudonym, userResponse.Pseudonym)
	}

	if userResponse.IsAdmin != user.IsAdmin {
		t.Errorf("Expected is_admin %t. Got %t\n", user.IsAdmin, userResponse.IsAdmin)
	}

	if userResponse.ID != user.ID {
		t.Errorf("Expected id %d. Got %d\n", user.ID, userResponse.ID)
	}
}

func createJSONString(user *model.User) string {
	return fmt.Sprintf(`{"username":"%s", "pseudonym": "%s", "password": "%s", "is_admin": %t}`, user.Username, user.Pseudonym, user.Password, user.IsAdmin)
}

func createXMLString(user *model.User) string {
	return fmt.Sprintf(`<User><username>%s</username><pseudonym>%s</pseudonym><password>%s</password><is_admin>%t</is_admin></User>`, user.Username, user.Pseudonym, user.Password, user.IsAdmin)
}
