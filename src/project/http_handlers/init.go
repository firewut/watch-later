package http_handlers

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/gorilla/securecookie"
	"net/http"
	"project/modules/config"
	"project/modules/log"
	"project/modules/messages"
	"runtime/debug"
	"time"
)

var (
	Config                   *config.Config
	MuxVars                  = make(map[string]string)
	Router                   *mux.Router
	Log                      *log.Log
	allowed_mulipart_methods = []string{
		"POST",
		"PUT",
		"PATCH",
	}
	s *securecookie.SecureCookie
)

func SetHashes() {
	hashKey := []byte(Config.HashKey)
	blockKey := []byte(Config.BlockKey)
	s = securecookie.New(hashKey, blockKey)
}

func HttpHandler(
	handler_function func(*Http),
) func(http.ResponseWriter, *http.Request) {
	return func(writer http.ResponseWriter, request *http.Request) {
		// Permanently write headers and return
		if request.Method == "OPTIONS" {
			writer.Header().Set("Access-Control-Allow-Methods", "GET")
			return
		}

		startTime := time.Now()
		_http := NewHttp()
		_http.startTime = startTime

		_http.initResponse(request, writer)
		if _http.Error == nil && _http.InternalError == nil {
			handler_function(_http)
		}
		_http.commitResponse()

		defer func(_http_passed *Http) {
			if Config.Mode != "development" {
				if x := recover(); x != nil {
					Log.Error(fmt.Sprintf("%[1]s %[2]s caught panic: %v\n %s\n\n", request.RemoteAddr, request.URL.RequestURI(), x, debug.Stack()))
					_http_passed.SetInternalError(messages.ErrInternalError)
				}
			}
		}(_http)
	}
}
