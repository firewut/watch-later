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
	CHECK_WATCH_LATER_PERIOD        = time.Minute * 5
	ISO_8601_DURATION_REGEX         = `^PT(?P<minutes>\d+)M(?P<seconds>\d+)S$`
	WATCH_LATER_PLAYLIST_NAME       = `Watch Later Service`
	WATCH_LATER_PLAYLIST_NAME_SHORT = `WLS`
	WATCH_LATER_PLAYLIST_REGEXP     = `^((Watch Later Service)|(Watch Later Service[[ ]{1}\d+]?))$`
)

// A Youtube's playlist
type Playlist struct {
	Id         string
	Title      string
	VideoItems Videos // < playlist videos
}

type Playlists []*Playlist

func (p Playlists) Len() int           { return len(p) }
func (p Playlists) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }
func (p Playlists) Less(i, j int) bool { return p[i].Title < p[j].Title }

// Video item from a playlist
type Video struct {
	Id       string
	Title    string
	Duration time.Duration
}

type Videos []*Video

func (v Videos) Len() int           { return len(v) }
func (v Videos) Swap(i, j int)      { v[i], v[j] = v[j], v[i] }
func (v Videos) Less(i, j int) bool { return v[i].Duration < v[j].Duration }

// Duration Playlists
type DurationPlaylist struct {
	playlistId string
	start      time.Duration
	end        time.Duration
}

func (dp *DurationPlaylist) checkMatches(d time.Duration) bool {
	return d <= dp.end && d >= dp.start
}

func newDurationPlaylists() (DurationPlaylists map[string]*DurationPlaylist) {
	DurationPlaylists = map[string]*DurationPlaylist{
		fmt.Sprintf("%s 0-3 minutes", WATCH_LATER_PLAYLIST_NAME_SHORT): &DurationPlaylist{
			start: time.Second,
			end:   time.Minute*3 + time.Second*59,
		},
		fmt.Sprintf("%s 4-7 minutes", WATCH_LATER_PLAYLIST_NAME_SHORT): &DurationPlaylist{
			start: time.Minute * 4,
			end:   time.Minute*7 + time.Second*59,
		},
		fmt.Sprintf("%s 8-10 minutes", WATCH_LATER_PLAYLIST_NAME_SHORT): &DurationPlaylist{
			start: time.Minute * 8,
			end:   time.Minute*10 + time.Second*59,
		},
		fmt.Sprintf("%s 10 minutes-12 hours", WATCH_LATER_PLAYLIST_NAME_SHORT): &DurationPlaylist{
			start: time.Minute * 10,
			end:   time.Hour*12 + time.Second*59,
		},
	}

	return
}

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
	headers.Set("X-Action", "Find")
	return headers
}

// this will process a response from `https://deform.io`
func processDeformResponse(response *http.Response, page_class interface{}) (response_json map[string]interface{}, paginated_response PaginatedResponse, err error) {
	body_data, _ := ioutil.ReadAll(response.Body)

	if page_class != nil {
		switch {
		case helpers.IsKind(page_class, reflect.ValueOf(PaginatedResponseResult{}).Kind()):
			paginated_response = &PaginatedResponseResult{}
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
	token_schema_sync_response, err := helpers.Put(
		t.GetCollectionURL(),
		getDeformJSONHeaders(getDeformHeaders(make(http.Header))),
		Config.Mode == "prod",
		map[string]interface{}{
			"_id":    "user_tokens",
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

	if err != nil {
		panic(err)
	}

	processed_token_schema_sync_response_response, _, _ := processDeformResponse(token_schema_sync_response, nil)
	if token_schema_sync_response.StatusCode == 200 ||
		token_schema_sync_response.StatusCode == 201 {
		// pass
	} else {
		Config.RavenClient.CaptureError(
			fmt.Errorf("%v. %v",
				processed_token_schema_sync_response_response["message"],
				processed_token_schema_sync_response_response["errors"],
			),
			map[string]string{},
		)
		panic(
			fmt.Sprintf(
				"%v",
				processed_token_schema_sync_response_response,
			),
		)
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
type PaginatedResponseResult struct {
	Links   map[string]string `json:"links"`
	Total   int64             `json:"total"`
	Page    int64             `json:"page"`
	Pages   int64             `json:"pages"`
	PerPage int64             `json:"per_page"`
	Tokens  []*Token          `json:"items"`
	Message string            `json:"message"`
	Errors  []interface{}     `json:"errors"`
}

func (t *PaginatedResponseResult) GetPage() int64 {
	return t.Page
}

func (t *PaginatedResponseResult) GetPages() int64 {
	return t.Pages
}

func (t *PaginatedResponseResult) GetPerPage() int64 {
	return t.PerPage
}

func (t *PaginatedResponseResult) GetTotal() int64 {
	return t.Total
}

func (t *PaginatedResponseResult) GetLinks() map[string]string {
	return t.Links
}

func (t *PaginatedResponseResult) GetResults() []interface{} {
	var slice_to_return []interface{}
	for _, token_ptr := range t.Tokens {
		slice_to_return = append(slice_to_return, token_ptr)
	}
	return slice_to_return
}
