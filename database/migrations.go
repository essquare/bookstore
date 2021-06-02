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

package database

import (
	"database/sql"
)

var schemaVersion = len(migrations)

// Order is important. Add new migrations at the end of the list.
var migrations = []func(tx *sql.Tx) error {
	func(tx *sql.Tx) (err error) {
		sql := `
			CREATE TABLE schema_version (
				version TEXT NOT NULL
			);
			
			CREATE TABLE users (
				user_id INTEGER PRIMARY KEY AUTOINCREMENT,
				username TEXT NOT NULL UNIQUE,
				password TEXT NOT NULL,
				pseudonym TEXT NOT NULL UNIQUE,
				is_admin INTEGER DEFAULT '0'
			);

			CREATE TABLE books (
				book_id INTEGER PRIMARY KEY AUTOINCREMENT,
				user_id INTEGER NOT NULL REFERENCES users(user_id) ON DELETE CASCADE ON UPDATE CASCADE,
				image_url TEXT NOT NULL,
				title TEXT NOT NULL,
				description TEXT NOT NULL,
				price INTEGER NOT NULL,
				UNIQUE (user_id, title)
			);
			
			`
		_, err = tx.Exec(sql)
		return err
	},
}
