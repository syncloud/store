package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
)

func postJSON(t *testing.T, h echo.HandlerFunc, body interface{}) (*httptest.ResponseRecorder, error) {
	b, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(string(b)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	e := echo.New()
	ctx := e.NewContext(req, rec)
	err := h(ctx)
	return rec, err
}
