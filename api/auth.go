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
	"net/http/httputil"
	"time"

	"bookstore/auth"

	log "github.com/sirupsen/logrus"
)

type tokenMsg struct {
	Token string `json:"token" xml:"token"`
}

func (h *handler) authenticate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		log.Error("[authenticate] Method not Post")
		renderResult(w, r, http.StatusBadRequest, strToObjectError("Authenticate should be Post"))
		return
	}
	err := r.ParseForm()
	if err != nil {
		log.Error("[authenticate] Could not parse form")
		renderResult(w, r, http.StatusBadRequest, strToObjectError("Could not parse parameters"))
		return
	}

	dump, _ := httputil.DumpRequest(r, false)
	log.Infof("%s", dump)
	username := r.Form.Get("username")
	password := r.Form.Get("password")

	if username == "" || password == "" {
		log.Error("[authenticate] Empty username or password")
		renderResult(w, r, http.StatusBadRequest, strToObjectError("Username or password empty"))
		return
	}

	err = h.store.CheckPassword(username, password)
	if err != nil {
		log.Error("[authenticate] Username or password not correct")
		renderResult(w, r, http.StatusBadRequest, strToObjectError("Credentials incorrect"))
		return
	}

	token, err := auth.JWTToken(username, time.Now().Add(tokenValidity))

	if err != nil {
		log.Errorf("[authenticate] Could not create token: %v", err)
		renderResult(w, r, http.StatusInternalServerError, strToObjectError("Internal Server Error"))
		return
	}

	renderResult(w, r, http.StatusOK, &tokenMsg{Token: token})
}
