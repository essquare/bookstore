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

import "encoding/xml"

// User represents a user in the system.
type User struct {
	XMLName   xml.Name `json:"-" xml:"User"`
	ID        int64    `json:"id" xml:"id,attr"`
	Username  string   `json:"username" xml:"username"`
	Password  string   `json:"-" xml:"-"`
	Pseudonym string   `json:"pseudonym" xml:"pseudonym"`
	IsAdmin   bool     `json:"is_admin" xml:"is_admin"`
}

type users []User

// Users represents a list of users.
type Users struct {
	users
}

func (u Users) List() []interface{} {
	b := make([]interface{}, len(u.users))
	for i := range u.users {
		b[i] = u.users[i]
	}
	return b
}

// NewUsers returns new Users struct
func NewUsers(users []User) *Users {
	return &Users{users: users}
}

// UserCreationRequest represents the request to create a user.
type UserCreationRequest struct {
	XMLName   xml.Name `json:"-" xml:"User"`
	Username  string   `json:"username" xml:"username"`
	Password  string   `json:"password" xml:"password"`
	Pseudonym string   `json:"pseudonym" xml:"pseudonym"`
	IsAdmin   bool     `json:"is_admin" xml:"is_admin"`
}

// UserModificationRequest represents the request to modify a user.
type UserModificationRequest struct {
	XMLName   xml.Name `json:"-" xml:"User"`
	Username  *string  `json:"username" xml:"username"`
	Password  *string  `json:"password" xml:"password"`
	Pseudonym *string  `json:"pseudonym" xml:"pseudonym"`
	IsAdmin   *bool    `json:"is_admin" xml:"is_admin"`
}

// Patch updates the User object with the modification request.
func (u *UserModificationRequest) Patch(user *User) {
	if u.Username != nil {
		user.Username = *u.Username
	}

	if u.Password != nil {
		user.Password = *u.Password
	}

	if u.IsAdmin != nil {
		user.IsAdmin = *u.IsAdmin
	}

	if u.Pseudonym != nil {
		user.Pseudonym = *u.Pseudonym
	}
}
