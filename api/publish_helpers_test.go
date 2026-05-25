package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/labstack/echo/v4"
)

type fakeMP struct {
	mu       sync.Mutex
	objects  map[string][]byte
	uploadId string
	parts    [][]*s3.CompletedPart
	getErr   error
}

func newFakeMP() *fakeMP { return &fakeMP{objects: map[string][]byte{}} }

func (f *fakeMP) Create(_ string) (string, error) { return "upload-1", nil }
func (f *fakeMP) PresignPart(k, u string, n int) (string, error) {
	return "https://s3.example/?p=" + u + "&n=" + itoa(n), nil
}
func (f *fakeMP) Complete(_, _ string, parts []*s3.CompletedPart) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.parts = append(f.parts, parts)
	return nil
}
func (f *fakeMP) Abort(_, _ string) error          { return nil }
func (f *fakeMP) HeadSize(k string) (int64, error) { return int64(len(f.objects[k])), nil }
func (f *fakeMP) Put(k string, b []byte, _ string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.objects[k] = b
	return nil
}
func (f *fakeMP) Get(k string) ([]byte, error) {
	if f.getErr != nil {
		return nil, f.getErr
	}
	b, ok := f.objects[k]
	if !ok {
		return nil, errors.New("nosuchkey")
	}
	return b, nil
}

func itoa(i int) string {
	const d = "0123456789"
	if i == 0 {
		return "0"
	}
	s := ""
	for i > 0 {
		s = string(d[i%10]) + s
		i /= 10
	}
	return s
}

type fakeCache struct{ refreshed bool }

func (f *fakeCache) Refresh() error { f.refreshed = true; return nil }

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
