package cmd

import (
	"encoding/json"
)

type httpError struct {
	Cause  error  `json:"-"`
	Detail string `json:"error"` // Message to the user
	Status int    `json:"-"`     // The http status code to return the user
}

func (e *httpError) Error() string {
	if e.Cause == nil {
		return e.Detail
	}
	return e.Detail + " : " + e.Cause.Error()
}

func (e *httpError) ResponseBody() ([]byte, error) {
	body, err := json.Marshal(e)
	if err != nil {
		return nil, err
	}
	return body, nil
}

func (e *httpError) ResponseHeaders() (int, map[string]string) {
	return e.Status, map[string]string{
		"Content-Type": "application/json; charset=utf-8",
	}
}

func newHTTPError(err error, status int, detail string) error {
	return &httpError{
		Cause:  err,
		Detail: detail,
		Status: status,
	}
}

type clientError interface {
	Error() string
	// ResponseBody returns response body.
	ResponseBody() ([]byte, error)
	// ResponseHeaders returns http status code and headers.
	ResponseHeaders() (int, map[string]string)
}
