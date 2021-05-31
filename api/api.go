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
	router.HandleFunc("/authenticate", handler.authenticate).Methods(http.MethodPost).Name("Authenticate")
	router.Handle("/users", middleware.handleToken(http.HandlerFunc(handler.listUsers))).Methods(http.MethodGet).Name("ListUsers")
	router.Handle("/users", middleware.handleToken(http.HandlerFunc(handler.createUser))).Methods(http.MethodPost).Name("CreateUser")
	router.Handle("/users/{userID:[0-9]+}", middleware.handleToken(http.HandlerFunc(handler.updateUser))).Methods(http.MethodPut).Name("UpdateUser")
	router.Handle("/users/{userID:[0-9]+}", middleware.handleToken(http.HandlerFunc(handler.deleteUser))).Methods(http.MethodDelete).Name("DeleteUser")
	router.Handle("/users/{userID:[0-9]+}", middleware.handleToken(http.HandlerFunc(handler.getUser))).Methods(http.MethodGet).Name("GetUser")

	router.Handle("/users/{userID:[0-9]+}/books", middleware.handleToken(http.HandlerFunc(handler.listUserBooks))).Methods(http.MethodGet).Name("ListUserBooks")
	router.Handle("/users/{userID:[0-9]+}/books", middleware.handleToken(http.HandlerFunc(handler.createUserBook))).Methods(http.MethodPost).Name("CreateUserBook")
	router.Handle("/users/{userID:[0-9]+}/books/{bookID:[0-9]+}", middleware.handleToken(http.HandlerFunc(handler.updateUserBook))).Methods(http.MethodPut).Name("UpdateUserBook")
	router.Handle("/users/{userID:[0-9]+}/books/{bookID:[0-9]+}", middleware.handleToken(http.HandlerFunc(handler.deleteUserBook))).Methods(http.MethodDelete).Name("DeleteUserBook")

	// router.HandleFunc("/books" handler.books).Methods("GET")
	router.Handle("/books/{bookID:[0-9]+}", http.HandlerFunc(handler.getBook)).Methods(http.MethodGet).Name("GetBook")
}
