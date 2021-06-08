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

package model

import (
	"encoding/xml"
)

const (
	DefaultBookSorting          = "title"
	DefaultBookSortingDirection = "desc"
)

type Book struct {
	XMLName     xml.Name `json:"-" xml:"book"`
	ID          int64    `json:"id" xml:"id,attr"`
	UserID      int64    `json:"user_id" xml:"user_id"`
	User        *User    `json:"user,omitempty" xml:"user"`
	Title       string   `json:"title" xml:"title"`
	Description string   `json:"description" xml:"description"`
	Price       int64    `json:"price" xml:"price"`
	ImageURL    string   `json:"image_url" xml:"image_url"`
}

// Books is a list of book
type Books struct {
	XMLName   xml.Name `json:"-" xml:"books"`
	Books []Book   `json:"-" xml:"book"`
}

// NewBooks returns new Books struct
func NewBooks(books []Book) *Books {
	return &Books{Books: books}
}

func (u *Books) List() []interface{} {
	b := make([]interface{}, len(u.Books))
	for i := range u.Books {
		b[i] = u.Books[i]
	}
	return b
}

func (u *Books) InternalList() interface{} {
	return &u.Books
}

// UserCreationRequest represents the request to create a user.
type BookCreationRequest struct {
	XMLName     xml.Name `json:"-" xml:"book"`
	Title       string   `json:"title" xml:"title"`
	Description string   `json:"description" xml:"description"`
	Price       int64    `json:"price" xml:"price"`
	ImageURL    string   `json:"image_url" xml:"image_url"`
}

// UserModificationRequest represents the request to modify a user.
type BookModificationRequest struct {
	XMLName     xml.Name `json:"-" xml:"book"`
	Title       *string  `json:"title" xml:"title"`
	Description *string  `json:"description" xml:"description"`
	Price       *int64   `json:"price" xml:"price"`
	ImageURL    *string  `json:"image_url" xml:"image_url"`
}

// Patch updates the User object with the modification request.
func (b *BookModificationRequest) Patch(book *Book) {
	if b.Title != nil {
		book.Title = *b.Title
	}

	if b.Description != nil {
		book.Description = *b.Description
	}

	if b.Price != nil {
		book.Price = *b.Price
	}

	if b.ImageURL != nil {
		book.ImageURL = *b.ImageURL
	}
}

type BookListingRequest struct {
	Title       *string
	Description *string
	MinPrice    *int64
	MaxPrice    *int64
	AutorID     *int64
}
