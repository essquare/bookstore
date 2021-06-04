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

func listUserBooks(t *testing.T, caller map[string]interface{}, expectedBooks []map[string]interface{}, contentType string) {
	var m model.Books

	r := NewRequest(caller, fmt.Sprintf("/users/%v/books", (caller)["id"]), http.MethodGet, nil, "Book", contentType, contentType)
	response := r.makeRequest(t)
	checkResponseCode(t, response.Code, http.StatusOK)

	r.unmarshal(t, response, &m)

	checkBooks(t, expectedBooks, &m)
}

func TestGeneralBookOperations(t *testing.T) {
	resetDatabase(t)
	admin := createDefaultAdmin(t)

	contentTypes := []string{contentXML, contentJSON, contentAlternateXML}

	for _, contentType := range contentTypes {
		book := map[string]interface{}{
			"title":       "The test book",
			"description": "More details for testing",
			"image_url":   "https://images.books/cover.jpg",
			"user_id":     admin["id"],
			"price":       int64(1995),
		}

		createBook(t, admin, &book, contentType)

		change := map[string]interface{}{
			"title":       "The final book",
			"description": "Some more description for this book",
			"price":       int64(1499),
		}

		updateBook(t, admin, &book, change, contentType)

		deleteBook(t, admin, &book, contentType)

		createBook(t, admin, &book, contentType)

		deleteBook(t, admin, &book, contentType)
	}
}

func TestListUserBooks(t *testing.T) {
	resetDatabase(t)

	admin := createDefaultAdmin(t)
	millerUser := createSimpleUser(t, "millerUser", "Michael Miller")
	brownUser := createSimpleUser(t, "brownUser", "Dan Brown")

	theMillersB := map[string]interface{}{
		"title":       "The Millers",
		"description": "Today we meet the Millers",
		"image_url":   "https://images.books/millers.jpg",
		"user_id":     millerUser["id"],
		"price":       int64(1995),
	}

	daVinciB := map[string]interface{}{
		"title":       "Da Vinci Code",
		"description": "Some spooky stuff",
		"image_url":   "https://images.books/vinci.jpg",
		"user_id":     brownUser["id"],
		"price":       int64(995),
	}

	infernoB := map[string]interface{}{
		"title":       "Inferno",
		"description": "More about symbolic stuff",
		"image_url":   "https://images.books/inferno.jpg",
		"user_id":     brownUser["id"],
		"price":       int64(2000),
	}

	allBooks := []map[string]interface{}{
		theMillersB,
		daVinciB,
		infernoB,
	}

	for _, book := range allBooks {
		createBook(t, admin, &book, contentJSON)
	}

	contentTypes := []string{contentJSON, contentXML, contentAlternateXML}

	for _, contentType := range contentTypes {
		listUserBooks(t, admin, nil, contentType)
	}

	for _, contentType := range contentTypes {
		expectedResult := []map[string]interface{}{
			theMillersB,
		}
		listUserBooks(t, millerUser, expectedResult, contentType)
	}

	for _, contentType := range contentTypes {
		expectedResult := []map[string]interface{}{
			infernoB,
			daVinciB,
		}
		listUserBooks(t, brownUser, expectedResult, contentType)
	}
}



