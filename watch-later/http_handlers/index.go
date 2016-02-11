package http_handlers

import (
	"bytes"
	"html/template"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"project/models"
	"project/modules/messages"
)

// Serve index page
func Index() func(w http.ResponseWriter, r *http.Request) {
	index_template, err := template.ParseFiles(filepath.Join(Config.TemplatesDir, "index.html"))
	if err != nil {
		panic(err)
	}

	return HttpHandler(func(h *Http) {
		var token_obj *models.Token

		// if we have a user with a cookie - redirect him to profile page
		if cookie, err := h.request.Cookie("profile-id"); err == nil {
			value := make(map[string]string)
			if err = s.Decode("profile-id", cookie.Value, &value); err == nil {
				// if we have a user with a cookie - check that he has a token active
				if token, err := models.GetTokenForProfile(value["profile-id"]); err == nil {
					token_obj = token
				} else if err == messages.ErrTokenNotFound {
					// pass
				} else {
					Log.Error(err)
					h.SetError(messages.ErrInternalError)
					return
				}
			}
		}

		h.no_commit_response = true
		index_template.ExecuteTemplate(
			h.writer, "index.html",
			map[string]interface{}{
				"token": token_obj,
			},
		)
	})
}

// Remove app from a list of authorized apps in user's interface
func Stop() func(w http.ResponseWriter, r *http.Request) {
	return HttpHandler(func(h *Http) {
		if cookie, err := h.request.Cookie("profile-id"); err == nil {
			value := make(map[string]string)
			if err = s.Decode("profile-id", cookie.Value, &value); err == nil {
				// if we have a user with a cookie - check that he has a token active
				if token, err := models.GetTokenForProfile(value["profile-id"]); err == nil {
					if token.CheckIsActive() {
						if err := token.Revoke(); err == nil {
							h.no_commit_response = true
							http.Redirect(h.writer, h.request, "/", http.StatusFound)
							return
						} else {
							Log.Error(err)
							h.SetError(err)
							return
						}
					}
				} else {
					Log.Error(err)
					h.SetError(messages.ErrInternalError)
					return
				}
			}
		}

		h.SetError(messages.ErrProfileNotFound)
	})
}

// Enable service for a profile
func Enable() func(w http.ResponseWriter, r *http.Request) {
	return HttpHandler(func(h *Http) {
		if cookie, err := h.request.Cookie("profile-id"); err == nil {
			value := make(map[string]string)
			if err = s.Decode("profile-id", cookie.Value, &value); err == nil {
				// if we have a user with a cookie - check that he has a token active
				if token, err := models.GetTokenForProfile(value["profile-id"]); err == nil {
					if !token.CheckIsActive() {
						if err = token.Enable(); err == nil {
							h.no_commit_response = true
							http.Redirect(h.writer, h.request, "/", http.StatusFound)
							return
						} else {
							Log.Error(err)
							h.SetError(err)
							return
						}
					}
				} else {
					Log.Error(err)
					h.SetError(messages.ErrInternalError)
					return
				}
			}
		}

		h.SetError(messages.ErrProfileNotFound)
	})
}

// Disable service for a profile
func Disable() func(w http.ResponseWriter, r *http.Request) {
	return HttpHandler(func(h *Http) {
		if cookie, err := h.request.Cookie("profile-id"); err == nil {
			value := make(map[string]string)
			if err = s.Decode("profile-id", cookie.Value, &value); err == nil {
				// if we have a user with a cookie - check that he has a token active
				if token, err := models.GetTokenForProfile(value["profile-id"]); err == nil {
					if token.CheckIsActive() {
						if err = token.Disable(); err == nil {
							h.no_commit_response = true
							http.Redirect(h.writer, h.request, "/", http.StatusFound)
							return
						} else {
							Log.Error(err)
							h.SetError(err)
							return
						}
					}
				} else {
					Log.Error(err)
					h.SetError(messages.ErrInternalError)
					return
				}
			}
		}

		h.SetError(messages.ErrProfileNotFound)
	})
}

// Serve profile page
func Profile() func(w http.ResponseWriter, r *http.Request) {
	return HttpHandler(func(h *Http) {
		if cookie, err := h.request.Cookie("profile-id"); err == nil {
			value := make(map[string]string)
			if err = s.Decode("profile-id", cookie.Value, &value); err == nil {
				// if we have a user with a cookie - check that he has a token active
				if token, err := models.GetTokenForProfile(value["profile-id"]); err == nil {
					h.SetResponse(token.Profile)
					return

				} else {
					Log.Error(err)
					h.SetError(messages.ErrInternalError)
					return
				}
			}
		}

		h.SetError(messages.ErrProfileNotFound)
	})
}

// Serve favicon
func Favicon() func(w http.ResponseWriter, r *http.Request) {
	template, err := ioutil.ReadFile(filepath.Join(Config.TemplatesDir, "static/favicon.png"))
	if err != nil {
		panic(err)
	}

	return HttpHandler(func(h *Http) {
		h.writeReader = true
		h.SetResponse(bytes.NewReader([]byte(template)))
	})
}

// Serve logo
func Logo() func(w http.ResponseWriter, r *http.Request) {
	template, err := ioutil.ReadFile(filepath.Join(Config.TemplatesDir, "static/logo.png"))
	if err != nil {
		panic(err)
	}

	return HttpHandler(func(h *Http) {
		h.writeReader = true
		h.SetResponse(bytes.NewReader([]byte(template)))
	})
}
