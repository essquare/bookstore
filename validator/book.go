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

package validator

import (
	"net/url"

	"bookstore.app/model"
	"bookstore.app/storage"
)

// ValidateBookCreation validates user creation with a password.
func ValidateBookCreation(store *storage.Storage, userID int64, request *model.BookCreationRequest) error {
	if request.Title == "" {
		return NewValidationError("book_mandatory_fields:title")
	}

	if store.BookForUserExists(userID, request.Title) {
		return NewValidationError("book_already_exists")
	}
	if request.ImageURL != "" {
		_, err := url.ParseRequestURI(request.ImageURL)
		if err != nil {
			return NewValidationError("invalid_book_fields:image_url")
		}
	}

	if request.Price < 0 {
		return NewValidationError("invalid_book_fields:price")
	}

	return nil
}

// ValidateBookModification validates user modifications.
func ValidateBookModification(store *storage.Storage, userID int64, bookID int64, changes *model.BookModificationRequest) error {
	if changes.Title != nil {
		if *changes.Title == "" {
			return NewValidationError("book_mandatory_fields:title")
		} else if store.BookWithSameTitle(userID, bookID, *changes.Title) {
			return NewValidationError("invalid_book_fields:title")
		}
	}

	if changes.Price != nil {
		if *changes.Price < 0 {
			return NewValidationError("invalid_book_fields:price")
		}
	}


	return nil
}
