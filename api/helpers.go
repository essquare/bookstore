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
	"encoding/json"
	"encoding/xml"
	"fmt"
	"net/http"
	"strconv"

	"bookstore/model"

	"github.com/elnormous/contenttype"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	"github.com/unrolled/render"
)

type errorMsg struct {
	XMLName      xml.Name `json:"-" xml:"Error"`
	ErrorMessage string   `json:"error_message" xml:"error_message"`
}

const (
	ApplicationJSON = "application/json"
	AppicationXML   = "application/xml"
	TextXML         = "text/xml"
)

type lister interface {
	List() []interface{}
}

type ListContainer interface {
	List() []interface{}
	InternalList() interface{}
}

// ContextKey represents a context key.
type ContextKey int

// List of context keys.
const (
	UserKey ContextKey = iota
	ContentTypeKey
	AcceptKey
)

func requestUser(r *http.Request) (*model.User, error) {
	if v := r.Context().Value(UserKey); v != nil {
		value, valid := v.(*model.User)
		if !valid {
			return nil, fmt.Errorf("value is not from type model.User %v", v)
		}

		return value, nil
	}

	return nil, fmt.Errorf("no value for key user in context")
}

func requestContentType(r *http.Request) (*contenttype.MediaType, error) {
	if v := r.Context().Value(ContentTypeKey); v != nil {
		value, valid := v.(*contenttype.MediaType)
		if !valid {
			return nil, fmt.Errorf("value is not from type contenttype.MediaType %v", v)
		}

		return value, nil
	}

	return nil, fmt.Errorf("no value for key ContentType in context")
}
func requestAccept(r *http.Request) (*contenttype.MediaType, error) {
	if v := r.Context().Value(AcceptKey); v != nil {
		value, valid := v.(*contenttype.MediaType)
		if !valid {
			return nil, fmt.Errorf("value is not from type contenttype.MediaType %v", v)
		}

		return value, nil
	}

	return nil, fmt.Errorf("no value for key Accept in context")
}
func errToObjectError(err error) *errorMsg {
	return &errorMsg{ErrorMessage: err.Error()}
}
func strToObjectError(err string) *errorMsg {
	return &errorMsg{ErrorMessage: err}
}

func queryInt64Param(r *http.Request, param string) (*int64, error) {
	vars := r.URL.Query()
	v, ok := vars[param]
	if !ok {
		return nil, nil
	}
	if len(v) > 1 {
		return nil, fmt.Errorf("more than one query parameter: %s", param)
	}
	value, err := strconv.ParseInt(v[0], 10, 64)
	if err != nil {
		return nil, err
	}

	if value < 0 {
		return nil, fmt.Errorf("number is negative")
	}

	return &value, nil
}

func queryStringParam(r *http.Request, param string) (*string, error) {
	vars := r.URL.Query()
	v, ok := vars[param]
	if !ok {
		return nil, nil
	}

	return &v[0], nil
}

func routeInt64Param(r *http.Request, param string) int64 {
	vars := mux.Vars(r)
	value, err := strconv.ParseInt(vars[param], 10, 64)
	if err != nil {
		return 0
	}

	if value < 0 {
		return 0
	}

	return value
}

func renderResult(w http.ResponseWriter, r *http.Request, status int, object interface{}) {
	render := render.New()
	accept, err := requestAccept(r)
	if err != nil {
		log.Errorf("[renderResult] No Accept in context: %v", err)
		render.Text(w, http.StatusInternalServerError, "Server Error")
		return
	}

	switch accept.Type + "/" + accept.Subtype {
	case ApplicationJSON:
		if o, ok := object.(lister); ok {
			render.JSON(w, status, o.List())
		} else {
			render.JSON(w, status, object)
		}
	case AppicationXML, TextXML:
		render.XML(w, status, object)
	default:
		log.Errorf("[renderResult] accept type unknow - should not happen - %v", accept.Type)
		render.Text(w, http.StatusNotFound, "Accept Type unknown")
		return
	}
}

func unmarshalRequestObject(w http.ResponseWriter, r *http.Request, object interface{}) error {
	contenttype, err := requestContentType(r)
	if err != nil {
		return err
	}
	switch contenttype.Type + "/" + contenttype.Subtype {
	case ApplicationJSON:
		return json.NewDecoder(r.Body).Decode(object)
	case AppicationXML, TextXML:
		return xml.NewDecoder(r.Body).Decode(object)
	}
	return fmt.Errorf("content type unknow - %s/%s", contenttype.Type, contenttype.Subtype)
}
