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

package validator

import (
	"bookstore/model"
	"bookstore/storage"
)

// ValidateUserCreationWithPassword validates user creation with a password.
func ValidateUserCreation(store *storage.Storage, request *model.UserCreationRequest) error {
	if request.Username == "" {
		return NewValidationError("user_mandatory_fields:username")
	}

	if request.Pseudonym == "" {
		return NewValidationError("user_mandatory_fields:pseudonym")
	}

	if store.UserExists(request.Username) {
		return NewValidationError("user_already_exists")
	}

	if err := validatePassword(request.Password); err != nil {
		return err
	}

	return nil
}

// ValidateUserModification validates user modifications.
func ValidateUserModification(store *storage.Storage, userID int64, changes *model.UserModificationRequest) error {
	if changes.Username != nil {
		if *changes.Username == "" {
			return NewValidationError("user_mandatory_fields:username")
		} else if store.AnotherUserExists(userID, *changes.Username) {
			return NewValidationError("user_already_exists")
		}
	}

	if changes.Password != nil {
		if err := validatePassword(*changes.Password); err != nil {
			return err
		}
	}

	return nil
}
func validatePassword(password string) error {
	if len(password) < 6 {
		return NewValidationError("password_min_length")
	}
	return nil
}
