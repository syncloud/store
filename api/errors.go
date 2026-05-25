package api

import "net/http"

type apiError struct {
	Status int
	Msg    string
}

func (e *apiError) Error() string { return e.Msg }

func badRequest(msg string) error  { return &apiError{Status: http.StatusBadRequest, Msg: msg} }
func unauthorized() error          { return &apiError{Status: http.StatusUnauthorized, Msg: "unauthorized"} }
func conflict(msg string) error    { return &apiError{Status: http.StatusConflict, Msg: msg} }
