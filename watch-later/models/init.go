package models

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"io/ioutil"
	"net/http"
	"project/modules/config"
	"project/modules/helpers"
	"project/modules/log"
	"reflect"
	"strings"
	"time"
)

const (
	CHECK_WATCH_LATER_PERIOD  = time.Minute * 5
	WATCH_LATER_PLAYLIST_NAME = `Watch Later Service`
)

var (
	Config *config.Config
	Log    *log.Log
	Router *mux.Router

	// schema for user token
	user_token_collection_schema = `{
		"type": "object",
		"properties": {
			"oauth2token": {
				"type": "object",
				"properties": {
					"access_token": {
						"type": "string"
					},
					"token_type": {
						"type": "string"
					},
					"refresh_token": {
						"type": "string"
					},
					"expiry": {
						"type": "datetime"
					}
				}
			},
			"latest_operation": {
				"type": "datetime"
			},
			"is_active": {
				"type": "boolean"
			},
			"watch_later_playlist_id": {
				"type": "string"
			},
			"profile": {
				"type": "object",
				"properties": {
					"id": {
						"type": "string"
					},
					"name": {
						"type": "string"
					},
					"link": {
						"type": "string"
					},
					"picture": {
						"type": "string"
					},
					"gender": {
						"type": "string"
					},
					"locale": {
						"type": "string"
					},
					"email": {
						"type": "string"
					},
					"verified_email": {
						"type": "boolean"
					}
				}
			}
		}
	}`
)

func getDeformHeaders(headers http.Header) http.Header {
	headers.Set("Authorization", fmt.Sprintf("Token %s", Config.DeformIOToken))
	return headers
}

func getDeformJSONHeaders(headers http.Header) http.Header {
	headers.Set("Content-Type", "application/json")
	return headers
}

func getDeformSearchHeaders(headers http.Header) http.Header {
	headers.Set("X-Action", "search")
	return headers
}

// this will process a response from `https://deform.io`
func processDeformResponse(response *http.Response, page_class interface{}) (response_json map[string]interface{}, paginated_response PaginatedResponse, err error) {
	body_data, _ := ioutil.ReadAll(response.Body)

	if page_class != nil {
		switch {
		case helpers.IsKind(page_class, reflect.ValueOf(TokensPaginatedResponse{}).Kind()):
			paginated_response = &TokensPaginatedResponse{}
		}
	}

	response_json = make(map[string]interface{})
	if strings.Contains(response.Header.Get("Content-Type"), "application/json") {
		if len(body_data) > 0 {
			// if a page is not paginated - fill a map of it
			err := json.Unmarshal(body_data, &response_json)
			if err != nil {
				// Try to parse it supposing it's a string
				string_response := ""
				err = json.Unmarshal(body_data, &string_response)
				if err == nil {
					response_json["string"] = string_response
					return response_json, paginated_response, err
				}
				array_map_response := []map[string]interface{}{}
				err = json.Unmarshal(body_data, &array_map_response)
				if err == nil {
					response_json["array_map"] = array_map_response
					return response_json, paginated_response, err
				}
				array_response := []interface{}{}
				err = json.Unmarshal(body_data, &array_response)
				if err == nil {
					response_json["array"] = array_response
					return response_json, paginated_response, err
				}
			}

			// if we passed an object to be filled with a paginated response
			if paginated_response != nil {
				// If we have a deal with a paginated page
				if _, total_ok := response_json["total"]; total_ok {
					err := json.Unmarshal(body_data, &paginated_response)
					if err != nil {
						return response_json, paginated_response, err
					}
				}
			}
		}
	}
	return
}

// Call it to sync schemas
func SyncSchemas() {
	// sync token collection schema
	t := NewToken(nil, nil)
	token_schema_sync_response, _ := helpers.Put(
		t.GetCollectionURL(),
		getDeformJSONHeaders(getDeformHeaders(make(http.Header))),
		map[string]interface{}{
			"id":     "user_tokens",
			"name":   "user tokens",
			"schema": user_token_collection_schema,
			"indexes": []map[string]interface{}{
				{
					"type":     "simple",
					"property": "oauth2.access_token",
				},
				{
					"type":     "simple",
					"property": "user_id",
				},
				{
					"type":     "simple",
					"property": "is_active",
				},
			},
		},
	)

	processed_token_schema_sync_response_response, _, _ := processDeformResponse(token_schema_sync_response, nil)
	if token_schema_sync_response.StatusCode == 200 ||
		token_schema_sync_response.StatusCode == 201 {
		// pass
	} else {
		Log.Error(processed_token_schema_sync_response_response["error"])
		panic(processed_token_schema_sync_response_response["error"])
	}
}

// If we have a paginated response
type PaginatedResponse interface {
	GetPage() int64
	GetPages() int64
	GetPerPage() int64
	GetTotal() int64
	GetLinks() map[string]string
	GetResults() []interface{}
}

// Response page from a `deform.io` will be parsed here
type TokensPaginatedResponse struct {
	Links   map[string]string `json:"links"`
	Total   int64             `json:"total"`
	Page    int64             `json:"page"`
	Pages   int64             `json:"page"`
	PerPage int64             `json:"page"`
	Tokens  []*Token          `json:"result"`
	Error   string            `json:"error"`
}

func (t *TokensPaginatedResponse) GetPage() int64 {
	return t.Page
}

func (t *TokensPaginatedResponse) GetPages() int64 {
	return t.Pages
}

func (t *TokensPaginatedResponse) GetPerPage() int64 {
	return t.PerPage
}

func (t *TokensPaginatedResponse) GetTotal() int64 {
	return t.Total
}

func (t *TokensPaginatedResponse) GetLinks() map[string]string {
	return t.Links
}

func (t *TokensPaginatedResponse) GetResults() []interface{} {
	var slice_to_return []interface{}
	for _, token_ptr := range t.Tokens {
		slice_to_return = append(slice_to_return, token_ptr)
	}
	return slice_to_return
}
