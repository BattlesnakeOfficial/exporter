package http

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/julienschmidt/httprouter"
	"github.com/stretchr/testify/require"
)

func testRequest(t *testing.T, method, url string, body io.Reader) (*http.Request, *httptest.ResponseRecorder) {
	clientRequest, err := http.NewRequest(method, url, body)
	require.NoError(t, err)

	recorder := httptest.NewRecorder()

	return clientRequest, recorder
}

func TestVersion(t *testing.T) {
	server := NewServer()

	os.Setenv("APP_VERSION", "1.2.3")
	defer os.Unsetenv("APP_VERSION")

	req, res := testRequest(t, "GET", "http://localhost/version", nil)

	server.Router.ServeHTTP(res, req)
	require.Equal(t, http.StatusOK, res.Code)
	require.Equal(t, "1.2.3", res.Body.String())
}

func TestHandlesPanic(t *testing.T) {
	server := NewServer()
	server.Router.GET("/fake/panic", httprouter.Handle(func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		panic("an unexpected error")
	}))

	req, res := testRequest(t, "GET", "http://localhost/fake/panic", nil)

	server.Router.ServeHTTP(res, req)
	require.Equal(t, http.StatusInternalServerError, res.Code)
}
