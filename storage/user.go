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
	"fmt"
	"strings"

	"bookstore/model"

	"golang.org/x/crypto/bcrypt"
)

// UserByUsername finds a user by the username.
func (s *Storage) UserByUsername(username string) (*model.User, error) {
	query := `
		SELECT
			user_id,
			username,
			is_admin,
			pseudonym
		FROM
			users
		WHERE
			username=LOWER($1)
	`
	return s.fetchUser(query, username)
}

func (s *Storage) UserByID(userID int64) (*model.User, error) {
	query := `
		SELECT
			user_id,
			username,
			is_admin,
			pseudonym
		FROM
			users
		WHERE
			user_id = $1
	`
	return s.fetchUser(query, userID)
}
func (s *Storage) fetchUser(query string, args ...interface{}) (*model.User, error) {
	var user model.User
	err := s.db.QueryRow(query, args...).Scan(
		&user.ID,
		&user.Username,
		&user.IsAdmin,
		&user.Pseudonym,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		return nil, fmt.Errorf(`store: unable to fetch user: %v`, err)
	}

	return &user, nil
}

// Users returns all users.
func (s *Storage) Users() (*model.Users, error) {
	query := `
		SELECT
			user_id,
			username,
			is_admin,
			pseudonym
		FROM
			users
		ORDER BY username ASC
	`
	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf(`store: unable to fetch users: %v`, err)
	}
	defer rows.Close()

	users := make([]model.User, 0)
	for rows.Next() {
		var user model.User
		err := rows.Scan(
			&user.ID,
			&user.Username,
			&user.IsAdmin,
			&user.Pseudonym,
		)

		if err != nil {
			return nil, fmt.Errorf(`store: unable to fetch users row: %v`, err)
		}

		users = append(users, user)
	}

	return model.NewUsers(users), nil
}

// CheckPassword validate the hashed password.
func (s *Storage) CheckPassword(username, password string) error {
	var hash string
	username = strings.ToLower(username)

	err := s.db.QueryRow("SELECT password FROM users WHERE username=$1", username).Scan(&hash)
	if err == sql.ErrNoRows {
		return fmt.Errorf(`store: unable to find this user: %s`, username)
	} else if err != nil {
		return fmt.Errorf(`store: unable to fetch user: %v`, err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)); err != nil {
		return fmt.Errorf(`store: invalid password for "%s" (%v)`, username, err)
	}

	return nil
}

func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// CreateUser creates a new user.
func (s *Storage) CreateUser(userCreationRequest *model.UserCreationRequest) (*model.User, error) {
	var hashedPassword string
	var err error
	if userCreationRequest.Password != "" {
		hashedPassword, err = hashPassword(userCreationRequest.Password)
		if err != nil {
			return nil, err
		}
	}

	query := `
		INSERT INTO users
			(username, password, is_admin, pseudonym)
		VALUES
			(LOWER($1), $2, $3, $4)
		RETURNING
			user_id,
			username,
			is_admin,
			pseudonym
	`

	var user model.User
	err = s.db.QueryRow(
		query,
		userCreationRequest.Username,
		hashedPassword,
		userCreationRequest.IsAdmin,
		userCreationRequest.Pseudonym,
	).Scan(
		&user.ID,
		&user.Username,
		&user.IsAdmin,
		&user.Pseudonym,
	)
	if err != nil {
		return nil, fmt.Errorf(`store: unable to create user %q: %v`, user.Username, err)
	}
	return &user, nil
}

// UpdateUser updates a user.
func (s *Storage) UpdateUser(user *model.User) error {
	if user.Password != "" {
		hashedPassword, err := hashPassword(user.Password)
		if err != nil {
			return err
		}

		query := `
			UPDATE users SET
				username=LOWER($1),
				password=$2,
				is_admin=$3,
				pseudonym=$4
			WHERE
				user_id=$5
		`

		_, err = s.db.Exec(
			query,
			user.Username,
			hashedPassword,
			user.IsAdmin,
			user.Pseudonym,
			user.ID,
		)
		if err != nil {
			return fmt.Errorf(`store: unable to update user: %v`, err)
		}
	} else {
		query := `
			UPDATE users SET
				username=LOWER($1),
				is_admin=$2,
				pseudonym=$3
			WHERE
				user_id=$4
		`

		_, err := s.db.Exec(
			query,
			user.Username,
			user.IsAdmin,
			user.Pseudonym,
			user.ID,
		)

		if err != nil {
			return fmt.Errorf(`store: unable to update user: %v`, err)
		}
	}

	return nil
}

func (s *Storage) DeleteUser(userID int64) error {
	_, err := s.db.Exec(`DELETE FROM users WHERE user_id=$1`, userID)
	return err
}

// UserExists checks if a user exists by using the given username.
func (s *Storage) UserExists(username string) bool {
	var result bool
	s.db.QueryRow(`SELECT true FROM users WHERE username=LOWER($1)`, username).Scan(&result)
	return result
}

func (s *Storage) UserWithPseudonymExists(pseudonym string) bool {
	var result bool
	s.db.QueryRow(`SELECT true FROM users WHERE pseudonym=$1`, pseudonym).Scan(&result)
	return result
}

// AnotherUserExists checks if another user exists with the given username.
func (s *Storage) AnotherUserExists(userID int64, username string) bool {
	var result bool
	s.db.QueryRow(`SELECT true FROM users WHERE user_id != $1 AND username=LOWER($2)`, userID, username).Scan(&result)
	return result
}
