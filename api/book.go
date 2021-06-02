// Copyright 2021 Gergan Penkov
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

func (h *handler) listBooks(w http.ResponseWriter, r *http.Request) {
	authorID, err := queryInt64Param(r, "author-id")
	if err != nil {
		log.Errorf("[ListBooks] Error reading query parameter: %v", err)
		renderResult(w, r, http.StatusBadRequest, errToObjectError(err))
		return
	}

	minPrice, err := queryInt64Param(r, "min-price")
	if err != nil {
		log.Errorf("[ListBooks] Error reading query parameter: %v", err)
		renderResult(w, r, http.StatusBadRequest, errToObjectError(err))
		return
	}

	maxPrice, err := queryInt64Param(r, "max-price")
	if err != nil {
		log.Errorf("[ListBooks] Error reading query parameter: %v", err)
		renderResult(w, r, http.StatusBadRequest, errToObjectError(err))
		return
	}

	title, err := queryStringParam(r, "title")
	if err != nil {
		log.Errorf("[ListBooks] Error reading query parameter: %v", err)
		renderResult(w, r, http.StatusBadRequest, errToObjectError(err))
		return
	}

	description, err := queryStringParam(r, "description")
	if err != nil {
		log.Errorf("[ListBooks] Error reading query parameter: %v", err)
		renderResult(w, r, http.StatusBadRequest, errToObjectError(err))
		return
	}

	search := &model.BookListingRequest{Title: title, Description: description, AutorID: authorID, MinPrice: minPrice, MaxPrice: maxPrice}

	err = validator.ValidateBookListing(*search)

	if err != nil {
		log.Errorf("[ListBooks] Validation Error: %v", err)
		renderResult(w, r, http.StatusBadRequest, errToObjectError(err))
		return
	}

	books, err := h.store.SearchBooks(*search)
	if err != nil {
		log.Errorf("[ListBooks] Error in loading the books from the database: %v", err)
		renderResult(w, r, http.StatusInternalServerError, strToObjectError("Server Error"))
		return
	}

	renderResult(w, r, http.StatusOK, books)
}

func (h *handler) getBook(w http.ResponseWriter, r *http.Request) {
	bookID := routeInt64Param(r, "bookID")

	book, err := h.store.BookByID(bookID)
	if err != nil {
		log.Errorf("[GetBook] Error in loading the book from the database: %v", err)
		renderResult(w, r, http.StatusInternalServerError, strToObjectError("Server Error"))
		return
	}
	if book == nil {
		log.Errorf("[GetBook] Book with id %d not found", bookID)
		renderResult(w, r, http.StatusNotFound, strToObjectError("Resource Not Found"))
		return
	}
	renderResult(w, r, http.StatusOK, book)
}

func (h *handler) listUserBooks(w http.ResponseWriter, r *http.Request) {
	_, err := requestUser(r)
	if err != nil {
		log.Errorf("[ListUserBooks] No User in context: %v", err)
		renderResult(w, r, http.StatusInternalServerError, strToObjectError("Server Error"))

		return
	}

	userID := routeInt64Param(r, "userID")

	user, err := h.store.UserByID(userID)
	if err != nil {
		log.Errorf("[ListUserBooks] Error in loading the user from the database: %v", err)
		renderResult(w, r, http.StatusInternalServerError, strToObjectError("Server Error"))
		return
	}

	if user == nil {
		log.Errorf("[ListUserBooks] User with id %d not found", userID)
		renderResult(w, r, http.StatusNotFound, strToObjectError("Resource Not Found"))
		return
	}

	books, err := h.store.UserBooks(userID)
	if err != nil {
		log.Errorf("[ListUserBooks] Error in loading the user books from the database: %v", err)
		renderResult(w, r, http.StatusInternalServerError, strToObjectError("Server Error"))
		return
	}

	renderResult(w, r, http.StatusOK, books)
}

func (h *handler) createUserBook(w http.ResponseWriter, r *http.Request) {
	ru, err := requestUser(r)
	if err != nil {
		log.Errorf("[CreateUserBook] No user in context: %v", err)
		renderResult(w, r, http.StatusInternalServerError, strToObjectError("Server Error"))
		return
	}

	userID := routeInt64Param(r, "userID")

	user, err := h.store.UserByID(userID)
	if err != nil {
		log.Errorf("[CreateUserBook] Error loading the user from the database: %v", err)
		renderResult(w, r, http.StatusInternalServerError, strToObjectError("Server Error"))
		return
	}

	if user == nil {
		log.Errorf("[CreateUserBook] User with id %d not found", userID)
		renderResult(w, r, http.StatusNotFound, strToObjectError("Resource Not Found"))
		return
	}

	if !ru.IsAdmin && user.ID != ru.ID {
		log.Errorf("[CreateUserBook] User with id %d tried to create book for another user", ru.ID)
		renderResult(w, r, http.StatusForbidden, strToObjectError("Access Forbidden"))
		return
	}

	var bookCreationRequest model.BookCreationRequest
	if err := unmarshalRequestObject(w, r, &bookCreationRequest); err != nil {
		log.Errorf("[CreateUserBook] JSON decoding error: %v", err)
		renderResult(w, r, http.StatusBadRequest, errToObjectError(err))
		return
	}

	if err := validator.ValidateBookCreation(h.store, userID, &bookCreationRequest); err != nil {
		log.Errorf("[CreateUserBook] Validation error: %v", err)
		renderResult(w, r, http.StatusBadRequest, errToObjectError(err))
		return
	}
	b, err := h.store.CreateBook(userID, &bookCreationRequest)
	if err != nil {
		log.Errorf("[CreateUserBook] Error in user book creation from the database: %v", err)
		renderResult(w, r, http.StatusInternalServerError, strToObjectError("Server Error"))
		return
	}
	renderResult(w, r, http.StatusCreated, b)
}

func (h *handler) updateUserBook(w http.ResponseWriter, r *http.Request) {
	ru, err := requestUser(r)
	if err != nil {
		log.Errorf("[UpdateUserBook] No user in context: %v", err)
		renderResult(w, r, http.StatusInternalServerError, strToObjectError("Server Error"))
		return
	}

	userID := routeInt64Param(r, "userID")
	bookID := routeInt64Param(r, "bookID")

	book, err := h.store.BookByIDAndUserID(userID, bookID)
	if err != nil {
		log.Errorf("[UpdateUserBook] Error loading the user from the database: %v", err)
		renderResult(w, r, http.StatusInternalServerError, strToObjectError("Server Error"))
		return
	}

	if book == nil {
		log.Errorf("[UpdateUserBook] book with id %d not found", bookID)
		renderResult(w, r, http.StatusNotFound, strToObjectError("Resource Not Found"))
		return
	}

	if !ru.IsAdmin && book.User.ID != ru.ID {
		log.Errorf("[UpdateUserBook] User with id %d tried to update another user's book", ru.ID)
		renderResult(w, r, http.StatusForbidden, strToObjectError("Access Forbidden"))
		return
	}

	var bookModificationRequest model.BookModificationRequest
	if err := unmarshalRequestObject(w, r, &bookModificationRequest); err != nil {
		log.Errorf("[UpdateUserBook] JSON decoding error: %v", err)
		renderResult(w, r, http.StatusBadRequest, errToObjectError(err))
		return
	}

	if err := validator.ValidateBookModification(h.store, userID, bookID, &bookModificationRequest); err != nil {
		log.Errorf("[UpdateUserBook] Validation error: %v", err)
		renderResult(w, r, http.StatusBadRequest, errToObjectError(err))
		return
	}

	bookModificationRequest.Patch(book)
	err = h.store.UpdateBook(book)
	if err != nil {
		log.Errorf("[UpdateUserBook] Error in user book update from the database: %v", err)
		renderResult(w, r, http.StatusInternalServerError, strToObjectError("Server Error"))
		return
	}
	renderResult(w, r, http.StatusOK, book)
}

func (h *handler) deleteUserBook(w http.ResponseWriter, r *http.Request) {
	ru, err := requestUser(r)
	if err != nil {
		log.Errorf("[DeleteUserBook] No user in context: %v", err)
		renderResult(w, r, http.StatusInternalServerError, strToObjectError("Server Error"))
		return
	}

	userID := routeInt64Param(r, "userID")
	bookID := routeInt64Param(r, "bookID")

	book, err := h.store.BookByIDAndUserID(userID, bookID)
	if err != nil {
		log.Errorf("[DeleteUserBook] Error loading the user from the database: %v", err)
		renderResult(w, r, http.StatusInternalServerError, strToObjectError("Server Error"))
		return
	}

	if book == nil {
		log.Errorf("[DeleteUserBook] book with id %d not found", bookID)
		renderResult(w, r, http.StatusNotFound, strToObjectError("Resource Not Found"))
		return
	}

	if !ru.IsAdmin && book.User.ID != ru.ID {
		log.Errorf("[DeleteUserBook] User with id %d tried to delete another user's book", ru.ID)
		renderResult(w, r, http.StatusForbidden, strToObjectError("Access Forbidden"))
		return
	}

	err = h.store.DeleteBook(bookID)
	if err != nil {
		log.Errorf("[DeleteUserBook] Error in user book deletion from the database: %v", err)
		renderResult(w, r, http.StatusInternalServerError, strToObjectError("Server Error"))
		return
	}
	renderResult(w, r, http.StatusOK, book)
}
