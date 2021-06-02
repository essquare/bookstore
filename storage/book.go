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

package storage

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"bookstore.app/model"
)

// BookQueryBuilder builds a SQL query to fetch entries.
type BookQueryBuilder struct {
	store      *Storage
	args       []interface{}
	conditions []string
	order      string
	direction  string
	limit      int
	offset     int
}

// NewEntryQueryBuilder returns a new EntryQueryBuilder.
func NewBookQueryBuilder(store *Storage) *BookQueryBuilder {
	return &BookQueryBuilder{
		store:      store,
		args:       []interface{}{},
		conditions: []string{},
	}
}

// WithOrder set the sorting order.
func (b *BookQueryBuilder) WithOrder(order string) *BookQueryBuilder {
	b.order = order
	return b
}

// WithDirection set the sorting direction.
func (b *BookQueryBuilder) WithDirection(direction string) *BookQueryBuilder {
	b.direction = direction
	return b
}

// WithLimit set the limit.
func (b *BookQueryBuilder) WithLimit(limit int) *BookQueryBuilder {
	if limit > 0 {
		b.limit = limit
	}
	return b
}

// WithUserID filter by user ID.
func (b *BookQueryBuilder) WithUserID(userID int64) *BookQueryBuilder {
	if userID > 0 {
		b.conditions = append(b.conditions, fmt.Sprintf("b.user_id = $%d", len(b.args)+1))
		b.args = append(b.args, userID)
	}
	return b
}

// WithBookID filter by book ID.
func (b *BookQueryBuilder) WithBookID(bookID int64) *BookQueryBuilder {
	if bookID > 0 {
		b.conditions = append(b.conditions, fmt.Sprintf("b.book_id = $%d", len(b.args)+1))
		b.args = append(b.args, bookID)
	}
	return b
}

// WithMinPrice filter by minimum price.
func (b *BookQueryBuilder) WithMinPrice(price int64) *BookQueryBuilder {
	if price >= 0 {
		b.conditions = append(b.conditions, fmt.Sprintf("b.price >= $%d", len(b.args)+1))
		b.args = append(b.args, price)
	}
	return b
}

// WithMaxPrice filter by maximum price.
func (b *BookQueryBuilder) WithMaxPrice(price int64) *BookQueryBuilder {
	if price >= 0 {
		b.conditions = append(b.conditions, fmt.Sprintf("b.price <= $%d", len(b.args)+1))
		b.args = append(b.args, price)
	}
	return b
}

// SearchTitle filter by title.
func (b *BookQueryBuilder) SearchTitle(title string) *BookQueryBuilder {
	if title != "" {
		b.conditions = append(b.conditions, fmt.Sprintf("b.title like $%d", len(b.args)+1))
		b.args = append(b.args, "%"+title+"%")
	}
	return b
}

// SearchDescription filter by description.
func (b *BookQueryBuilder) SearchDescription(description string) *BookQueryBuilder {
	if description != "" {
		b.conditions = append(b.conditions, fmt.Sprintf("b.description like $%d", len(b.args)+1))
		b.args = append(b.args, "%"+description+"%")
	}
	return b
}

// WithOffset set the offset.
func (b *BookQueryBuilder) WithOffset(offset int) *BookQueryBuilder {
	if offset > 0 {
		b.offset = offset
	}
	return b
}

func (b *BookQueryBuilder) buildCondition() string {
	if len(b.conditions) == 0 {
		return "1"
	}
	return strings.Join(b.conditions, " AND ")
}

func (b *BookQueryBuilder) buildSorting() string {
	var parts []string

	if b.order != "" {
		parts = append(parts, fmt.Sprintf(`ORDER BY %s`, b.order))
	}

	if b.direction != "" {
		parts = append(parts, b.direction)
	}

	if b.limit > 0 {
		parts = append(parts, fmt.Sprintf(`LIMIT %d`, b.limit))
	}

	if b.offset > 0 {
		parts = append(parts, fmt.Sprintf(`OFFSET %d`, b.offset))
	}

	return strings.Join(parts, " ")
}

// GetBooks returns a list of books that match the condition.
func (e *BookQueryBuilder) GetBooks() (*model.Books, error) {
	query := `
		SELECT
			b.book_id,
			b.user_id,
			b.title,
			b.description,
			b.price,
			b.image_url,
			u.username,
			u.is_admin,
			u.pseudonym
		FROM
			books b
		LEFT JOIN
			users u ON u.user_id=b.user_id
		WHERE %s %s
	`

	condition := e.buildCondition()
	sorting := e.buildSorting()
	query = fmt.Sprintf(query, condition, sorting)

	rows, err := e.store.db.Query(query, e.args...)
	if err != nil {
		return nil, fmt.Errorf("unable to get entries: %v", err)
	}
	defer rows.Close()

	entries := make([]model.Book, 0)
	for rows.Next() {
		var book model.Book
		// var iconID sql.NullInt64
		// var tz string

		book.User = &model.User{}

		err := rows.Scan(
			&book.ID,
			&book.UserID,
			&book.Title,
			&book.Description,
			&book.Price,
			&book.ImageURL,
			&book.User.Username,
			&book.User.IsAdmin,
			&book.User.Pseudonym,
		)

		if err != nil {
			return nil, fmt.Errorf("unable to fetch entry row: %v", err)
		}

		book.User.ID = book.UserID
		entries = append(entries, book)
	}

	return model.NewBooks(entries), nil
}

// GetBook returns a single book that match the condition.
func (b *BookQueryBuilder) GetBook() (*model.Book, error) {
	b.limit = 1
	books, err := b.GetBooks()
	if err != nil {
		return nil, err
	}

	if len(books.List()) != 1 {
		return nil, nil
	}

	book := books.List()[0].(model.Book)
	return &book, nil
}

// Books returns all books.
func (s *Storage) Books() (*model.Books, error) {
	builder := NewBookQueryBuilder(s)
	builder.WithOrder(model.DefaultBookSorting)
	builder.WithDirection(model.DefaultBookSortingDirection)
	return builder.GetBooks()
}

// Search Books search books.
func (s *Storage) SearchBooks(search model.BookListingRequest) (*model.Books, error) {
	builder := NewBookQueryBuilder(s)
	builder.WithOrder(model.DefaultBookSorting)
	builder.WithDirection(model.DefaultBookSortingDirection)
	if search.AutorID != nil {
		builder.WithUserID(*search.AutorID)
	}
	if search.Title != nil {
		builder.SearchTitle(*search.Title)
	}
	if search.Description != nil {
		builder.SearchDescription(*search.Description)
	}
	if search.MinPrice != nil {
		builder.WithMinPrice(*search.MinPrice)
	}
	if search.MaxPrice != nil {
		builder.WithMaxPrice(*search.MaxPrice)
	}

	return builder.GetBooks()
}

// BookByID returns a book by the ID.
func (s *Storage) BookByID(bookID int64) (*model.Book, error) {
	builder := NewBookQueryBuilder(s)
	builder.WithBookID(bookID)
	book, err := builder.GetBook()

	switch {
	case errors.Is(err, sql.ErrNoRows):
		return nil, nil
	case err != nil:
		return nil, fmt.Errorf(`store: unable to fetch book #%d: %v`, bookID, err)
	}

	return book, nil
}

// BookByIDAndUserID returns a book by the ID and User ID.
func (s *Storage) BookByIDAndUserID(userID, bookID int64) (*model.Book, error) {
	builder := NewBookQueryBuilder(s)
	builder.WithUserID(userID)
	builder.WithBookID(bookID)
	book, err := builder.GetBook()

	switch {
	case errors.Is(err, sql.ErrNoRows):
		return nil, nil
	case err != nil:
		return nil, fmt.Errorf(`store: unable to fetch book #%d: %v`, bookID, err)
	}

	return book, nil
}

func (s *Storage) UserBooks(userID int64) (*model.Books, error) {
	builder := NewBookQueryBuilder(s)
	builder.WithOrder(model.DefaultBookSorting)
	builder.WithDirection(model.DefaultBookSortingDirection)
	builder.WithUserID(userID)
	return builder.GetBooks()
}

// BookWithSameTitle checks if another book with a given title for the user exists
func (s *Storage) BookWithSameTitle(userID int64, bookID int64, title string) bool {
	var result bool
	s.db.QueryRow(`SELECT true FROM books WHERE book_id = $1 AND user_id = $2 AND title = $3`, userID, bookID, title).Scan(&result)
	return result
}

// BookForUserExists checks if another book with a given title for the user exists
func (s *Storage) BookForUserExists(userID int64, title string) bool {
	var result bool
	s.db.QueryRow(`SELECT true FROM books WHERE user_id = $1 AND title = $2`, userID, title).Scan(&result)
	return result
}

// CreateBook creates a new book.
func (s *Storage) CreateBook(userID int64, bookCreationRequest *model.BookCreationRequest) (*model.Book, error) {
	var err error

	user, err := s.UserByID(userID)
	if err != nil {
		return nil, err
	}
	query := `
		INSERT INTO books
			(user_id, title, description, price, image_url)
		VALUES
			($1, $2, $3, $4, $5)
		RETURNING
			book_id,
			user_id,
			title,
			description,
			price,
			image_url
	`

	var book model.Book
	err = s.db.QueryRow(
		query,
		userID,
		bookCreationRequest.Title,
		bookCreationRequest.Description,
		bookCreationRequest.Price,
		bookCreationRequest.ImageURL,
	).Scan(
		&book.ID,
		&book.UserID,
		&book.Title,
		&book.Description,
		&book.Price,
		&book.ImageURL,
	)
	if err != nil {
		return nil, fmt.Errorf(`store: unable to create book %s: %v`, book.Title, err)
	}
	book.User = user
	return &book, nil
}

// UpdateBook updates a book.
func (s *Storage) UpdateBook(book *model.Book) error {
	query := `
			UPDATE books SET
				title=$1,
				description=$2,
				price=$3,
				image_url=$4
			WHERE
				book_id=$5
		`

	_, err := s.db.Exec(
		query,
		book.Title,
		book.Description,
		book.Price,
		book.ImageURL,
		book.ID,
	)
	if err != nil {
		return fmt.Errorf(`store: unable to update book: %v`, err)
	}

	return nil
}

func (s *Storage) DeleteBook(bookID int64) error {
	_, err := s.db.Exec(`DELETE FROM books WHERE book_id=$1`, bookID)
	return err
}
