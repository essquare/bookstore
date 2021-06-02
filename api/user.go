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

package api

import (
	"net/http"

	"bookstore.app/model"
	"bookstore.app/validator"
	log "github.com/sirupsen/logrus"
)

func (h *handler) getUser(w http.ResponseWriter, r *http.Request) {
	_, err := requestUser(r)
	if err != nil {
		log.Errorf("[GetUser] No User in context: %v", err)
		renderResult(w, r, http.StatusInternalServerError, strToObjectError("Server Error"))

		return
	}

	userID := routeInt64Param(r, "userID")

	user, err := h.store.UserByID(userID)
	if err != nil {
		log.Errorf("[GetUser] Error in loading the user from the database: %v", err)
		renderResult(w, r, http.StatusInternalServerError, strToObjectError("Server Error"))
		return
	}

	if user == nil {
		log.Errorf("[GetUser] User with id %d not found", userID)
		renderResult(w, r, http.StatusNotFound, strToObjectError("Resource Not Found"))
		return
	}

	renderResult(w, r, http.StatusOK, user)
}

func (h *handler) listUsers(w http.ResponseWriter, r *http.Request) {
	_, err := requestUser(r)
	if err != nil {
		log.Errorf("[ListUsers] No user in context: %v", err)
		renderResult(w, r, http.StatusInternalServerError, strToObjectError("Server Error"))
		return
	}
	users, err := h.store.Users()
	if err != nil {
		log.Errorf("[ListUsers] Error in listing users from the database: %v", err)
		renderResult(w, r, http.StatusInternalServerError, strToObjectError("Server Error"))
		return
	}
	renderResult(w, r, http.StatusOK, users)
}

func (h *handler) createUser(w http.ResponseWriter, r *http.Request) {
	ru, err := requestUser(r)
	if err != nil {
		log.Errorf("[CreateUser] No user in context: %v", err)
		renderResult(w, r, http.StatusInternalServerError, strToObjectError("Server Error"))
		return
	}
	if !ru.IsAdmin {
		log.Errorf("[CreateUser] User %s is not admin", ru.Username)
		renderResult(w, r, http.StatusUnauthorized, strToObjectError("Unauthorized"))
		return
	}

	var userCreationRequest model.UserCreationRequest
	if err = unmarshalRequestObject(w, r, &userCreationRequest); err != nil {
		log.Errorf("[CreateUser] JSON decoding error: %v", err)
		renderResult(w, r, http.StatusBadRequest, errToObjectError(err))
		return
	}

	if err := validator.ValidateUserCreation(h.store, &userCreationRequest); err != nil {
		log.Errorf("[CreateUser] Validation error: %v", err)
		renderResult(w, r, http.StatusBadRequest, errToObjectError(err))
		return
	}
	u, err := h.store.CreateUser(&userCreationRequest)
	if err != nil {
		log.Errorf("[CreateUser] Error in user creation from the database: %v", err)
		renderResult(w, r, http.StatusInternalServerError, strToObjectError("Server Error"))
	}
	renderResult(w, r, http.StatusCreated, u)
}

func (h *handler) updateUser(w http.ResponseWriter, r *http.Request) {
	ru, err := requestUser(r)
	if err != nil {
		log.Errorf("[UpdateUser] No user in context: %v", err)
		renderResult(w, r, http.StatusInternalServerError, strToObjectError("Server Error"))
		return
	}

	userID := routeInt64Param(r, "userID")

	originalUser, err := h.store.UserByID(userID)
	if err != nil {
		log.Errorf("[UpdateUser] Error loading the user from the database: %v", err)
		renderResult(w, r, http.StatusInternalServerError, strToObjectError("Server Error"))
		return
	}

	if originalUser == nil {
		log.Errorf("[UpdateUser] User with id %d not found", userID)
		renderResult(w, r, http.StatusNotFound, strToObjectError("Resource Not Found"))
		return
	}

	var userModificationRequest model.UserModificationRequest
	if err = unmarshalRequestObject(w, r, &userModificationRequest); err != nil {
		log.Errorf("[UpdateUser] JSON decoding error: %v", err)
		renderResult(w, r, http.StatusBadRequest, errToObjectError(err))
		return
	}

	if !ru.IsAdmin {
		if originalUser.ID != ru.ID {
			log.Errorf("[UpdateUser] User with id %d tried to change another user", ru.ID)
			renderResult(w, r, http.StatusForbidden, strToObjectError("Access Forbidden"))
			return
		}

		if userModificationRequest.IsAdmin != nil && *userModificationRequest.IsAdmin {
			log.Errorf("[UpdateUser] User with id %d tried to become an admin", ru.ID)
			renderResult(w, r, http.StatusBadRequest, strToObjectError("Normal users could not change their permissions"))
			return
		}
	}

	if validationErr := validator.ValidateUserModification(h.store, originalUser.ID, &userModificationRequest); validationErr != nil {
		log.Errorf("[CreateUser] Validation error: %v", err)
		renderResult(w, r, http.StatusBadRequest, errToObjectError(err))
		return
	}

	userModificationRequest.Patch(originalUser)
	if err = h.store.UpdateUser(originalUser); err != nil {
		log.Errorf("[CreateUser] Error in user creation from the database: %v", err)
		renderResult(w, r, http.StatusInternalServerError, strToObjectError("Server Error"))
		return
	}

	renderResult(w, r, http.StatusOK, originalUser)
}

func (h *handler) deleteUser(w http.ResponseWriter, r *http.Request) {
	ru, err := requestUser(r)
	if err != nil {
		log.Errorf("[DeleteUser] No user in context: %v", err)
		renderResult(w, r, http.StatusInternalServerError, strToObjectError("Server Error"))
		return
	}

	if !ru.IsAdmin {
		log.Errorf("[DeleteUser] User with id %d tried to delete another user", ru.ID)
		renderResult(w, r, http.StatusForbidden, strToObjectError("Access Forbidden"))
		return
	}

	userID := routeInt64Param(r, "userID")
	userDelete, err := h.store.UserByID(userID)
	if err != nil {
		log.Errorf("[DeleteUser] Error in loading the user from the database: %v", err)
		renderResult(w, r, http.StatusInternalServerError, strToObjectError("Server Error"))
		return
	}

	if userDelete == nil {
		log.Errorf("[DeleteUser] User with id %d not found", userID)
		renderResult(w, r, http.StatusNotFound, strToObjectError("Resource Not Found"))
		return
	}

	if ru.ID == userDelete.ID {
		log.Errorf("[DeleteUser] User with id %d tried to delete him-/herself", ru.ID)
		renderResult(w, r, http.StatusBadRequest, strToObjectError("Deleting own account not possible"))
		return
	}

	err = h.store.DeleteUser(userID)
	if err != nil {
		log.Errorf("[DeleteUser] Error in deleting the user from the database: %v", err)
		renderResult(w, r, http.StatusInternalServerError, strToObjectError("Server Error"))
		return
	}

	renderResult(w, r, http.StatusNoContent, nil)
}
