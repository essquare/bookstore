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
	"time"

	"bookstore.app/storage"

	"github.com/gorilla/mux"
)

type handler struct {
	store *storage.Storage
}

const tokenValidity = 15 * time.Minute

// Serve declares API routes for the application.
func Serve(router *mux.Router, store *storage.Storage) {

	handler := &handler{store}

	middleware := newMiddleware(store)

	router.Use(middleware.handleMediaTypes)

	usersRoute := router.PathPrefix("/users").Subrouter()
	booksRoute := router.PathPrefix("/books").Subrouter()

	usersRoute.Use(middleware.handleToken)

	router.HandleFunc("/authenticate", handler.authenticate).Methods(http.MethodPost).Name("Authenticate")
	usersRoute.HandleFunc("", handler.listUsers).Methods(http.MethodGet).Name("ListUsers")
	usersRoute.HandleFunc("", handler.createUser).Methods(http.MethodPost).Name("CreateUser")
	usersRoute.HandleFunc("/{userID:[0-9]+}", handler.updateUser).Methods(http.MethodPut).Name("UpdateUser")
	usersRoute.HandleFunc("/{userID:[0-9]+}", handler.deleteUser).Methods(http.MethodDelete).Name("DeleteUser")
	usersRoute.HandleFunc("/{userID:[0-9]+}", handler.getUser).Methods(http.MethodGet).Name("GetUser")

	usersRoute.HandleFunc("/{userID:[0-9]+}/books", handler.listUserBooks).Methods(http.MethodGet).Name("ListUserBooks")
	usersRoute.HandleFunc("/{userID:[0-9]+}/books", handler.createUserBook).Methods(http.MethodPost).Name("CreateUserBook")
	usersRoute.HandleFunc("/{userID:[0-9]+}/books/{bookID:[0-9]+}", handler.updateUserBook).Methods(http.MethodPut).Name("UpdateUserBook")
	usersRoute.HandleFunc("/{userID:[0-9]+}/books/{bookID:[0-9]+}", handler.deleteUserBook).Methods(http.MethodDelete).Name("DeleteUserBook")

	booksRoute.HandleFunc("", handler.listBooks).Methods(http.MethodGet).Name("ListBooks")
	booksRoute.HandleFunc("/{bookID:[0-9]+}", handler.getBook).Methods(http.MethodGet).Name("GetBook")
}
