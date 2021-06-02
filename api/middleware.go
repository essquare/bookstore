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
	"context"
	"net/http"

	jwtmiddleware "github.com/auth0/go-jwt-middleware"
	"github.com/elnormous/contenttype"
	log "github.com/sirupsen/logrus"

	"bookstore.app/auth"
	"bookstore.app/storage"
	"github.com/form3tech-oss/jwt-go"
)

type middleware struct {
	store         *storage.Storage
	jwtMiddleware *jwtmiddleware.JWTMiddleware
}

func newMiddleware(s *storage.Storage) *middleware {
	jwtMiddleware := jwtmiddleware.New(jwtmiddleware.Options{
		ValidationKeyGetter: func(token *jwt.Token) (interface{}, error) {
			return auth.HMACKey(), nil
		},
		// When set, the middleware verifies that tokens are signed with the specific signing algorithm
		// If the signing method is not constant the ValidationKeyGetter callback can be used to implement additional checks
		// Important to avoid security issues described here: https://auth0.com/blog/critical-vulnerabilities-in-json-web-token-libraries/
		SigningMethod: auth.SigningMethod(),
		UserProperty:  "token",
	})
	return &middleware{s, jwtMiddleware}
}

func (m *middleware) handleMediaTypes(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mediaType, err := contenttype.GetMediaType(r)
		if err != nil {
			log.Errorf("[Middleware][MediaTypes] Content-Type could not be parsed: %v", err)
			http.Error(w, "Erronous Content-Type", http.StatusBadRequest)
			return
		}
		log.Info("Media type:", mediaType.String())

		availableMediaTypes := []contenttype.MediaType{
			contenttype.NewMediaType(ApplicationJSON),
			contenttype.NewMediaType(AppicationXML),
			contenttype.NewMediaType(TextXML),
		}

		accepted, extParameters, err := contenttype.GetAcceptableMediaType(r, availableMediaTypes)
		if err != nil {
			log.Errorf("[Middleware][MediaTypes] Accept could not be parsed: %v", err)
			http.Error(w, "Erronous Accept", http.StatusBadRequest)
			return

		}
		log.Info("Accepted media type:", accepted.String(), "extension parameters:", extParameters)

		ctx := r.Context()
		ctx = context.WithValue(ctx, ContentTypeKey, &mediaType)
		ctx = context.WithValue(ctx, AcceptKey, &accepted)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
func (m *middleware) handleToken(next http.Handler) http.Handler {

	return m.jwtMiddleware.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Context().Value("token")
		claims, ok := token.(*jwt.Token).Claims.(jwt.MapClaims)
		if !ok {
			log.Error("[Middleware][HandleToken] Token could not be cast to MapClaims")
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		sub, ok := claims["sub"].(string)
		if !ok {
			log.Error("[Middleware][HandleToken] sub could not be cast to string")
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		user, err := m.store.UserByUsername(sub)
		if err != nil {
			log.Errorf("[Middleware][HandleToken] Problem loading the user from the database: %v", err)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		if user == nil {
			log.Errorf("[Middleware][HandleToken] No user found with user name: %s", sub)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// If we get here, everything worked and we can set the
		// user property in context.
		newRequest := r.WithContext(context.WithValue(r.Context(), UserKey, user))
		// Update the current request with the new context information.
		*r = *newRequest
		next.ServeHTTP(w, r)
	}))
}
