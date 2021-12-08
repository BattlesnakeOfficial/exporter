package http

import (
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/BattlesnakeOfficial/exporter/fixtures"
	"github.com/stretchr/testify/require"
)

func TestVersion(t *testing.T) {
	server := NewServer()

	os.Setenv("APP_VERSION", "1.2.3")
	defer os.Unsetenv("APP_VERSION")

	req, res := fixtures.TestRequest(t, "GET", "http://localhost/version", nil)

	server.router.ServeHTTP(res, req)
	require.Equal(t, http.StatusOK, res.Code)
	require.Equal(t, "1.2.3", res.Body.String())
}

func TestHandleGIFGame_NotFound(t *testing.T) {
	server := NewServer()

	engineServer := fixtures.StubEngineServer(func(res http.ResponseWriter, req *http.Request) {
		res.WriteHeader(http.StatusNotFound)
	})

	req, res := fixtures.TestRequest(t, "GET", "http://localhost/games/EXAMPLE/gif", nil)
	query := req.URL.Query()
	query.Set("engine_url", engineServer.URL)
	req.URL.RawQuery = query.Encode()

	server.router.ServeHTTP(res, req)
	require.Equal(t, http.StatusNotFound, res.Code)
}

func TestHandleGIFGame_Success(t *testing.T) {
	fixtures.TestInRootDir()
	server := NewServer()

	engineServer := fixtures.StubEngineServer(func(res http.ResponseWriter, req *http.Request) {
		if strings.HasSuffix(req.URL.Path, "/frames") {
			_, _ = res.Write([]byte(fixtures.ExampleGameFramesResponse))
		} else {
			_, _ = res.Write([]byte(fixtures.ExampleGameResponse))
		}
	})

	req, res := fixtures.TestRequest(t, "GET", "http://localhost/games/EXAMPLE/gif", nil)
	query := req.URL.Query()
	query.Set("engine_url", engineServer.URL)
	req.URL.RawQuery = query.Encode()

	server.router.ServeHTTP(res, req)

	require.Equal(t, http.StatusOK, res.Code)
	require.Equal(t, "image/gif", res.Result().Header.Get("Content-Type"))
}
