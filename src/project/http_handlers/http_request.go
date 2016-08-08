package http_handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"project/modules/messages"
	"strings"
	"time"
)

type Http struct {
	startTime             time.Time
	InternalError         error
	Error                 error
	status_code           int
	immutable_status_code bool
	no_commit_response    bool

	// Core
	request  *http.Request
	writer   http.ResponseWriter
	response interface{}

	vars map[string]string

	// Helpers
	decorators  []string
	writeFile   bool
	writeReader bool
}

func NewHttp() (h *Http) {
	h = &Http{
		vars:                  make(map[string]string),
		status_code:           500,
		immutable_status_code: false,
		writeFile:             false,
		writeReader:           false,
	}
	return
}

// Should be called **BEFORE** SetStatusCode
func (h *Http) SetError(err error) {
	if error_struct, ok := err.(*messages.ErrorStruct); ok {
		switch error_struct.GetFormat() {
		case messages.ERROR_OBJECT_NOT_FOUND:
			h.SetStatusCode(404)
		case messages.ERROR_INTERNAL_ERROR:
			h.SetStatusCode(500)
		default:
			h.SetStatusCode(422)
		}
	} else {
		h.SetStatusCode(422)
	}
	h.Error = err
	h.SetResponse(err.Error())
}

func (h *Http) SetInternalError(err error) {
	h.SetStatusCode(500)
	h.InternalError = err
	h.SetResponse(err.Error())
}

// Set status code
func (h *Http) SetStatusCode(status_code int) {
	if !h.immutable_status_code {
		h.status_code = status_code
		h.immutable_status_code = true
	}
}

func (h *Http) SetResponse(data interface{}) {
	h.response = data
}

// 204 HAS NO RESPONSE BODY :D
func (h *Http) SetResponseDeletedObject() {
	h.SetStatusCode(204)
}

func (h *Http) SetResponseCreatedObject(data interface{}) {
	h.SetStatusCode(201)
	h.SetResponse(data)
}

func (h *Http) commitResponse() {
	if !h.no_commit_response {
		h.SetStatusCode(200)
		if h.writeReader {
			h.WriteReader()
		} else {
			h.WriteJson()
		}
	}
}

func (h *Http) initResponse(request *http.Request, writer http.ResponseWriter) {
	h.writer = writer
	h.request = request
}

func (h *Http) WriteReader() {
	if reader, ok := h.response.(*bytes.Reader); ok {
		http.ServeContent(h.writer, h.request, "", time.Now().UTC(), reader)
	} else {
		h.SetError(messages.ErrFileNotFound)
		h.WriteJson()
	}
}

func (h *Http) WriteJson() {
	var data interface{}

	if h.Error != nil || h.InternalError != nil {
		data = messages.ErrorStructured(h.response)
	} else {
		data = messages.MessageStructured(h.response)
	}

	if json_raw, err := json.Marshal(data); err != nil {
		h.SetInternalError(err)
	} else {
		h.writer.Header().Set("Access-Control-Allow-Origin", "*")
		h.writer.Header().Set("Content-Type", "application/json;charset=utf-8")
		h.writer.Header().Set("Content-Length", fmt.Sprintf("%d", len(json_raw)))
		h.writer.WriteHeader(h.status_code)

		if !strings.EqualFold(h.request.Method, "head") {
			h.writer.Write(json_raw)
		}
	}
}
