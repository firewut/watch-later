package messages

import (
	"encoding/json"
	"fmt"
	"strings"
)

type ErrorStruct struct {
	format            string
	args              []interface{}
	additionalMessage string
}

func (e *ErrorStruct) MarshalJSON() ([]byte, error) {
	error_json_mapped := map[string]string{
		"error": e.Error(),
	}
	return json.Marshal(error_json_mapped)
}

func (e *ErrorStruct) UnmarshalJSON(b []byte) error {
	return nil
}

func NewError(format string, args ...interface{}) (err *ErrorStruct) {
	err = &ErrorStruct{
		format: format,
		args:   args,
	}
	return
}

func (e *ErrorStruct) GetFormat() string {
	return e.format
}

func (e *ErrorStruct) GetMessage() string {
	errors_strings := make([]string, 0)
	for _, arg := range e.args {
		errors_strings = append(errors_strings, fmt.Sprintf("%v", arg))
	}
	return strings.Join(errors_strings, "\n")
}

func (e *ErrorStruct) AppendMessage(msg string) {
	e.additionalMessage = msg
}

func (e *ErrorStruct) Error() string {
	s := ""
	if len(e.additionalMessage) == 0 {
		s = fmt.Sprintf(e.format, e.args...)
	} else {
		s = fmt.Sprintf("%s\n%s",
			fmt.Sprintf(e.format, e.args...),
			e.additionalMessage,
		)
	}
	return s
}

type Err struct {
	Error interface{} `json:"error"`
}

type Msg struct {
	Result interface{} `json:"result"`
}

func ErrorStructured(result interface{}) Err {
	return Err{Error: result}
}

func MessageStructured(result interface{}) Msg {
	return Msg{Result: result}
}

const (
	// If text will match - http handler will fail in **switch**

	// General
	ERROR_FORBIDDEN            = "forbidden"
	ERROR_INTERNAL_ERROR       = "internal error"
	ERROR_INTERNAL_ERROR_MSG   = "internal error. %s"
	ERROR_BAD_REQUEST          = "bad request"
	ERROR_UNPROCESSABLE_ENTITY = "%s"
	ERROR_RETRY                = "Retry. %s"

	// Object existance
	ERROR_OBJECT_NOT_FOUND = "%s not found"
)

var (
	ErrInternalError = NewError(ERROR_INTERNAL_ERROR)
	ErrForbidden     = NewError(ERROR_FORBIDDEN)
	ErrBadRequest    = NewError(ERROR_BAD_REQUEST)
	ErrNotFound      = NewError(ERROR_OBJECT_NOT_FOUND)

	ErrTaskNotFound    = NewError(ERROR_OBJECT_NOT_FOUND, "Task")
	ErrFileNotFound    = NewError(ERROR_OBJECT_NOT_FOUND, "File")
	ErrProfileNotFound = NewError(ERROR_OBJECT_NOT_FOUND, "Profile")
	ErrTokenNotFound   = NewError(ERROR_OBJECT_NOT_FOUND, "Token")

	MsgSuccess = "Success"
)
