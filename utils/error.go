package utils

import (
	"bytes"
	"fmt"
)

const (
	ECONFLICT = "409" // action cannot be performed
	EINTERNAL = "500" // internal error
	EINVALID  = "422" // validation failed
	ENOTFOUND = "404" // entity does not exist
)

type Error struct {
	error
	// Machine-readable error code.
	Code string `json:"code"`

	// Human-readable message.
	Message string `json:"message"`

	// Nested error.
	Err error `json:"-"`
}

func (e *Error) Error() string {
	var buf bytes.Buffer

	if e.Err != nil {
		buf.WriteString(e.Message + " : " + e.Err.Error())
	} else {
		if e.Code != "" {
			fmt.Fprintf(&buf, "<%s> ", e.Code)
		}
		buf.WriteString(e.Message)
	}
	return buf.String()
}

func ErrorCode(err error) string {
	if err == nil {
		return ""
	} else if e, ok := err.(*Error); ok && e.Code != "" {
		return e.Code
	} else if ok && e.Err != nil {
		return ErrorCode(e.Err)
	}
	return EINTERNAL
}

func NewError(code string, message string, err error) *Error {
	appError := &Error{
		Code:    code,
		Message: message,
		Err:     err,
	}

	return appError
}
