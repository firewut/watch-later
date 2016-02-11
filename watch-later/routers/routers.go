package routers

import (
	"github.com/gorilla/mux"
	"project/http_handlers"
	"project/modules/config"
)

func NewRouter(config *config.Config) (s *mux.Router) {
	s = mux.NewRouter()

	mainHandler := s.PathPrefix("/").Subrouter()
	mainHandler.HandleFunc("/", http_handlers.Index()).Methods("GET").Name("Index")
	mainHandler.HandleFunc("/login", http_handlers.OAuth2Redirect()).Methods("GET").Name("OAuth2Redirect")
	mainHandler.HandleFunc("/logout", http_handlers.Logout()).Methods("GET").Name("Logout")
	mainHandler.HandleFunc("/oauth2callback", http_handlers.OAuth2Callback()).Methods("GET").Name("OAuth2Callback")
	mainHandler.HandleFunc("/profile", http_handlers.Profile()).Methods("GET").Name("Profile")
	mainHandler.HandleFunc("/stop", http_handlers.Stop()).Methods("GET").Name("Stop")
	mainHandler.HandleFunc("/favicon.png", http_handlers.Favicon()).Methods("GET").Name("Favicon")
	mainHandler.HandleFunc("/logo.png", http_handlers.Logo()).Methods("GET").Name("Logo")

	mainHandler.HandleFunc("/enable", http_handlers.Enable()).Methods("GET").Name("Enable")
	mainHandler.HandleFunc("/disable", http_handlers.Disable()).Methods("GET").Name("Disable")

	// staticHandler := s.PathPrefix("/static/").Subrouter()

	config.Router = s
	return
}
