package http_handlers

import (
	"golang.org/x/oauth2"
	"io/ioutil"
	"net/http"
	"project/models"
	"project/modules/messages"
)

// Logging out - unset cookie
func Logout() func(w http.ResponseWriter, r *http.Request) {
	return HttpHandler(func(h *Http) {
		if cookie, err := h.request.Cookie("profile-id"); err == nil {
			value := make(map[string]string)
			if err = s.Decode("profile-id", cookie.Value, &value); err == nil {
				// if we have a user with a cookie - check that he has a token active
				if _, err := models.GetTokenForProfile(value["profile-id"]); err == nil {
					if encoded, err := s.Encode("profile-id",
						map[string]string{"profile-id": ""},
					); err == nil {
						cookie := &http.Cookie{
							Name:  "profile-id",
							Value: encoded,
							Path:  "/",
						}
						http.SetCookie(h.writer, cookie)
					}
				} else if err == messages.ErrTokenNotFound {
					// pass
				} else {
					Config.RavenClient.CaptureError(
						err,
						map[string]string{
							"profileId": value["profile-id"],
						},
					)
					h.no_commit_response = true
					http.Redirect(
						h.writer,
						h.request,
						"/error",
						http.StatusFound,
					)
					return
				}
			}
		}

		http.Redirect(h.writer, h.request, "/", http.StatusFound)
		h.no_commit_response = true
	})
}

func OAuth2Redirect() func(w http.ResponseWriter, r *http.Request) {
	return HttpHandler(func(h *Http) {
		// do not execute h.commitResponse()
		h.no_commit_response = true

		// This check prevents the "/" handler from handling all requests by default
		if h.request.URL.Path != "/login" {
			http.NotFound(h.writer, h.request)

			return
		}

		cfg := models.GetYoutubeConfig()
		url := cfg.AuthCodeURL("state", oauth2.AccessTypeOffline)

		http.Redirect(h.writer, h.request, url, http.StatusFound)
	})
}

func OAuth2Callback() func(w http.ResponseWriter, r *http.Request) {
	return HttpHandler(func(h *Http) {
		err_from_google := h.request.FormValue("error")
		if len(err_from_google) > 0 {
			http.Redirect(h.writer, h.request, "/", http.StatusFound)
			return
		}

		code := h.request.FormValue("code")

		cfg := models.GetYoutubeConfig()
		oauth2token, err := cfg.Exchange(oauth2.NoContext, code)

		if err != nil {
			Config.RavenClient.CaptureError(
				err,
				map[string]string{},
			)
			h.no_commit_response = true
			http.Redirect(
				h.writer,
				h.request,
				"/error",
				http.StatusFound,
			)
			return
		}

		if oauth2token.Valid() {
			client := cfg.Client(oauth2.NoContext, oauth2token)
			_ = client

			response, err := http.Get("https://www.googleapis.com/oauth2/v2/userinfo?access_token=" + oauth2token.AccessToken)
			defer response.Body.Close()
			contents, err := ioutil.ReadAll(response.Body)
			if err != nil {
				Config.RavenClient.CaptureError(
					err,
					map[string]string{},
				)
				h.no_commit_response = true
				http.Redirect(
					h.writer,
					h.request,
					"/error",
					http.StatusFound,
				)
				return
			}

			profile, err := models.NewProfileFromResponse(contents)
			if err != nil {
				Config.RavenClient.CaptureError(
					err,
					map[string]string{},
				)
				h.no_commit_response = true
				http.Redirect(
					h.writer,
					h.request,
					"/error",
					http.StatusFound,
				)
				return
			}

			// set a cookie
			value := map[string]string{
				"profile-id": profile.Id,
			}
			if encoded, err := s.Encode("profile-id", value); err == nil {
				cookie := &http.Cookie{
					Name:  "profile-id",
					Value: encoded,
					Path:  "/",
				}
				http.SetCookie(h.writer, cookie)
			}

			// Check that we DO not have a token with that .Profile.Id
			var token *models.Token
			token, err = models.GetTokenForProfile(profile.Id)
			if err == nil {
				OAuth2TokenDiff := make(map[string]interface{})
				// if we have a refresh token
				if len(oauth2token.RefreshToken) > 0 {
					OAuth2TokenDiff["refresh_token"] = oauth2token.RefreshToken
				}
				if len(oauth2token.TokenType) > 0 {
					OAuth2TokenDiff["token_type"] = oauth2token.TokenType
				}
				if len(oauth2token.AccessToken) > 0 {
					OAuth2TokenDiff["access_token"] = oauth2token.AccessToken
				}
				if !oauth2token.Expiry.IsZero() {
					OAuth2TokenDiff["expiry"] = oauth2token.Expiry
				}
				token.Patch(map[string]interface{}{
					"oauth2token": OAuth2TokenDiff,
				})
			} else {
				token = models.NewToken(oauth2token, profile)
				token.Save()
			}

			go models.ProcessWatchLater(client, token)

			// do not execute h.commitResponse()
			h.no_commit_response = true
			http.Redirect(h.writer, h.request, "/", http.StatusFound)
		} else {

			Log.StdOut.Info.Println("oauth2token is invalid")
		}
	})
}
