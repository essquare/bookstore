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
	"fmt"
	"net/http"
	"testing"

	"bookstore.app/model"
)

func createBook(t *testing.T, caller map[string]interface{}, book *map[string]interface{}, contentType string) {
	var m model.Book
	r := NewRequest(caller, fmt.Sprintf("/users/%d/books", (*book)["user_id"]), http.MethodPost, *book, "Book", contentType, contentType)
	response := r.makeRequest(t)

	checkResponseCode(t, response.Code, http.StatusCreated)

	r.unmarshal(t, response, &m)
	(*book)["id"] = m.ID
	checkBook(t, *book, &m)

	getBook(t, caller, book, contentType)

}

func updateBook(t *testing.T, caller map[string]interface{}, book *map[string]interface{}, change map[string]interface{}, contentType string) {
	var m model.Book
	r := NewRequest(caller, fmt.Sprintf("/users/%d/books/%v", (*book)["user_id"], (*book)["id"]), http.MethodPut, change, "Book", contentType, contentType)
	response := r.makeRequest(t)
	checkResponseCode(t, response.Code, http.StatusOK)

	r.unmarshal(t, response, &m)

	for key, value := range change {
		(*book)[key] = value
	}

	checkBook(t, *book, &m)

	getBook(t, caller, book, contentType)

}

func deleteBook(t *testing.T, caller map[string]interface{}, book *map[string]interface{}, contentType string) {
	r := NewRequest(caller, fmt.Sprintf("/users/%d/books/%v", (*book)["user_id"], (*book)["id"]), http.MethodDelete, nil, "Book", contentType, contentType)
	response := r.makeRequest(t)
	checkResponseCode(t, response.Code, http.StatusNoContent)

	getNonExistentBook(t, caller, book, contentType)
}

func getNonExistentBook(t *testing.T, caller map[string]interface{}, book *map[string]interface{}, contentType string) {
	r := NewRequest(caller, fmt.Sprintf("/books/%v", (*book)["id"]), http.MethodGet, nil, "Book", contentType, contentType)
	response := r.makeRequest(t)
	checkResponseCode(t, response.Code, http.StatusNotFound)
}

func getBook(t *testing.T, caller map[string]interface{}, book *map[string]interface{}, contentType string) {
	var m model.Book

	r := NewRequest(caller, fmt.Sprintf("/books/%v", (*book)["id"]), http.MethodGet, nil, "Book", contentType, contentType)
	response := r.makeRequest(t)
	checkResponseCode(t, response.Code, http.StatusOK)

	r.unmarshal(t, response, &m)

	checkBook(t, *book, &m)
}

func TestGeneralBookOperations(t *testing.T) {
	resetDatabase(t)
	admin := createDefaultAdmin(t)

	contentTypes := []string{contentXML, contentJSON, contentAlternateXML}

	for _, contentType := range contentTypes {

		book := map[string]interface{}{
			"title":  "The test book",
			"description": "More details for testing",
			"image_url":  "https://images.books/cover.jpg",
			"user_id":  admin["id"],
			"price":  int64(1995),
		}

		createBook(t, admin, &book, contentType)

		change := map[string]interface{}{
			"title": "The final book",
			"description": "Some more description for this book",
			"price": int64(1499),
		}

		updateBook(t, admin, &book, change, contentType)

		deleteBook(t, admin, &book, contentType)

		createBook(t, admin, &book, contentType)

		deleteBook(t, admin, &book, contentType)
	}
}
