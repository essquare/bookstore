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

package auth

import (
	"crypto/rand"
	"time"

	"github.com/form3tech-oss/jwt-go"
)

var hmacKey []byte
var signingMethod *jwt.SigningMethodHMAC

func init() {
	hmacKey = generateRandomBytes(64)
	signingMethod = jwt.SigningMethodHS512
}

func generateRandomBytes(size int) []byte {
	b := make([]byte, size)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}

	return b
}

func JWTToken(username string, validity time.Time) (string, error) {
	claims := &jwt.StandardClaims{
		ExpiresAt: validity.Unix(),
		Issuer:    "bookstore",
		IssuedAt:  time.Now().Unix(),
		NotBefore: time.Now().Unix(),
		Subject:   username,
	}

	token := jwt.NewWithClaims(signingMethod, claims)

	return token.SignedString(hmacKey)
}

func HMACKey() []byte {
	return hmacKey
}

func SigningMethod() *jwt.SigningMethodHMAC {
	return signingMethod
}
