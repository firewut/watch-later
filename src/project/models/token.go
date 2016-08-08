package models

import (
	"fmt"
	"golang.org/x/oauth2"
	// "google.golang.org/api/youtube/v3"
	// "io/ioutil"
	"net/http"
	"project/modules/helpers"
	"project/modules/messages"
	"reflect"
	"time"
)

// User Token class
type Token struct {
	Id          string        `json:"_id"`
	IsActive    bool          `json:"is_active"`
	OAuth2Token *oauth2.Token `json:"oauth2token"`
	Profile     *Profile      `json:"profile"`

	// id of a playlist named `watch later`
	// in future user can rename it - we need to post videos there
	LatestOperation time.Time `json:"latest_operation"`
}

// Get new token instance
func NewToken(oauth2token *oauth2.Token, profile *Profile) *Token {
	t := &Token{
		OAuth2Token: oauth2token,
		IsActive:    true,
		Profile:     profile,
	}
	return t
}

func (t *Token) LatestOperationWasPerformed() string {
	if t.LatestOperation.IsZero() {
		return "just now"
	}

	return fmt.Sprintf("%.0f minutes ago", time.Now().UTC().Sub(t.LatestOperation).Minutes())
}

// Check that token is active
func (t *Token) CheckIsActive() bool {
	if t != nil {
		return t.IsActive && t.Profile != nil
	}

	return false
}

// Get a token by it's profile id
func GetTokenForProfile(profile_id string) (token *Token, err error) {
	var tokens []*Token
	tmp_token := NewToken(nil, nil)
	tokens, _, err = tmp_token.FilterTokens(1, map[string]interface{}{
		"filter": map[string]interface{}{
			"profile.id": profile_id,
		},
	})

	if err == nil {
		if len(tokens) > 0 {
			// think that we have 1 token 1 user cookie
			token = tokens[0]
		} else {
			err = messages.ErrTokenNotFound
		}
	}
	return
}

// Get active tokens
func GetActiveTokens(page int64) (tokens []*Token, next_page int64, err error) {
	tmp_token := NewToken(nil, nil)
	return tmp_token.FilterTokens(page, map[string]interface{}{
		"filter": map[string]interface{}{
			"is_active": true,
		},
	})
}

// Filter tokens
func (t *Token) FilterTokens(page int64, payload map[string]interface{}) (tokens []*Token, next_page int64, err error) {
	tokens = make([]*Token, 0)

	response, _ := helpers.Post(
		t.GetDocumentsURL(page),
		getDeformSearchHeaders(
			getDeformJSONHeaders(
				getDeformHeaders(
					make(http.Header),
				),
			),
		),
		Config.Mode == "prod",
		payload,
	)

	processed_response, paginated_tokens, _ := processDeformResponse(response, PaginatedResponseResult{})

	if response.StatusCode == 200 {
		// If this page contains results
		if paginated_tokens.GetTotal() > 0 {
			// if this page is not last
			if paginated_tokens.GetPage() < paginated_tokens.GetPages() {
				next_page = page + 1
			}

			results := paginated_tokens.GetResults()
			if len(results) > 0 {
				// if we deal with a page which contains tokens
				if helpers.IsKind(results, reflect.Slice) {
					slice := reflect.ValueOf(results)
					for i := 0; i < slice.Len(); i++ {
						token_item_interface := slice.Index(i).Interface()

						if token, ok := token_item_interface.(*Token); ok {
							tokens = append(
								tokens,
								token,
							)
						}
					}
				}
			}

		}
	} else {
		Config.RavenClient.CaptureError(
			fmt.Errorf("%v %v", processed_response["errors"], processed_response["message"]),
			map[string]string{},
		)
		err = messages.ErrInternalError
	}

	return
}

// Get a token document url
func (t *Token) GetDocumentURL() string {
	return fmt.Sprintf(
		"%s/collections/user_tokens/documents/%s/",
		Config.DeformIOProject,
		t.Id,
	)
}

// Get tokens documents collection url
func (t *Token) GetDocumentsURL(page int64) string {
	per_page := 100
	return fmt.Sprintf(
		"%s/collections/user_tokens/documents/?page=%d&per_page=%d",
		Config.DeformIOProject,
		page,
		per_page,
	)
}

// Get tokens collection url
func (t *Token) GetCollectionURL() string {
	return fmt.Sprintf(
		"%s/collections/user_tokens/",
		Config.DeformIOProject,
	)
}

// Save a token
func (t *Token) Save() (err error) {
	response, err := helpers.Put(
		t.GetDocumentURL(),
		getDeformJSONHeaders(
			getDeformHeaders(
				make(http.Header),
			),
		),
		Config.Mode == "prod",
		t,
	)

	processed_response, _, _ := processDeformResponse(response, nil)
	switch response.StatusCode {
	case 200, 201:
	default:
		Config.RavenClient.CaptureError(
			fmt.Errorf("%v %v", processed_response["errors"], processed_response["message"]),
			map[string]string{
				"Token": t.Id,
				"Email": t.Profile.Email,
			},
		)
		return messages.NewError(
			messages.ERROR_UNPROCESSABLE_ENTITY,
			"An error occurred",
		)
	}

	return nil
}

// Patch a token
func (t *Token) Patch(payload map[string]interface{}) (err error) {
	response, err := helpers.Patch(
		t.GetDocumentURL(),
		getDeformJSONHeaders(
			getDeformHeaders(
				make(http.Header),
			),
		),
		Config.Mode == "prod",
		payload,
	)

	processed_response, _, _ := processDeformResponse(response, nil)
	switch response.StatusCode {
	case 200:
	default:
		Config.RavenClient.CaptureError(
			fmt.Errorf("%v %v", processed_response["errors"], processed_response["message"]),
			map[string]string{
				"Token": t.Id,
				"Email": t.Profile.Email,
			},
		)
		return messages.NewError(
			messages.ERROR_UNPROCESSABLE_ENTITY,
			"An error occurred",
		)
	}

	return nil
}

// Delete a token
func (t *Token) Delete() (err error) {
	response, err := helpers.Delete(
		t.GetDocumentURL(),
		getDeformJSONHeaders(
			getDeformHeaders(
				make(http.Header),
			),
		),
		Config.Mode == "prod",
	)

	processed_response, _, _ := processDeformResponse(response, nil)

	switch response.StatusCode {
	case 204:
	default:
		Config.RavenClient.CaptureError(
			fmt.Errorf("%v %v", processed_response["errors"], processed_response["message"]),
			map[string]string{
				"Token": t.Id,
				"Email": t.Profile.Email,
			},
		)
		return messages.NewError(
			messages.ERROR_UNPROCESSABLE_ENTITY,
			"An error occurred",
		)
	}

	return nil
}

// Revoking a token
func (t *Token) Revoke() (err error) {
	response, err := http.Get("https://accounts.google.com/o/oauth2/revoke?token=" + t.OAuth2Token.AccessToken)
	defer response.Body.Close()

	if response.StatusCode == 200 {
		go t.Delete()
	} else {
		Config.RavenClient.CaptureError(
			err,
			map[string]string{
				"Token": t.Id,
				"Email": t.Profile.Email,
			},
		)
		return err
	}

	return
}

// Disabling a token
func (t *Token) Disable() (err error) {
	t.IsActive = false
	response, err := helpers.Patch(
		t.GetDocumentURL(),
		getDeformJSONHeaders(
			getDeformHeaders(
				make(http.Header),
			),
		),
		Config.Mode == "prod",
		map[string]interface{}{
			"is_active": false,
		},
	)

	processed_response, _, _ := processDeformResponse(response, nil)
	switch response.StatusCode {
	case 200:
	default:
		Config.RavenClient.CaptureError(
			fmt.Errorf("%v %v", processed_response["errors"], processed_response["message"]),
			map[string]string{
				"Token": t.Id,
				"Email": t.Profile.Email,
			},
		)
		return messages.NewError(
			messages.ERROR_UNPROCESSABLE_ENTITY,
			"An error occurred",
		)
	}

	return nil
}

// Enabling a token
func (t *Token) Enable() (err error) {
	t.IsActive = true
	response, err := helpers.Patch(
		t.GetDocumentURL(),
		getDeformJSONHeaders(
			getDeformHeaders(
				make(http.Header),
			),
		),
		Config.Mode == "prod",
		map[string]interface{}{
			"is_active": true,
		},
	)

	processed_response, _, _ := processDeformResponse(response, nil)
	switch response.StatusCode {
	case 200:
	default:
		Config.RavenClient.CaptureError(
			fmt.Errorf("%v %v", processed_response["errors"], processed_response["message"]),
			map[string]string{
				"Token": t.Id,
				"Email": t.Profile.Email,
			},
		)
		return messages.NewError(
			messages.ERROR_UNPROCESSABLE_ENTITY,
			"An error occurred",
		)
	}

	return nil
}
