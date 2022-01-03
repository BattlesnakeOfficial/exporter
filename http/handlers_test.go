package http

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/BattlesnakeOfficial/exporter/fixtures"
	"github.com/stretchr/testify/require"
)

func TestHandleVersion(t *testing.T) {
	server := NewServer()

	os.Setenv("APP_VERSION", "1.2.3")
	defer os.Unsetenv("APP_VERSION")

	req, res := fixtures.TestRequest(t, "GET", "http://localhost/version", nil)

	server.router.ServeHTTP(res, req)
	require.Equal(t, http.StatusOK, res.Code)
	require.Equal(t, "1.2.3", res.Body.String())
}

func TestHandleAvatar_BadRequest(t *testing.T) {
	server := NewServer()

	// Fow now we're careful to not construct tests that will pull .svg resources from media
	badRequestPaths := []string{
		"/garbage", // Invalid pattern

		"/1x1.svg",       // Invalid dimension
		"/1x9.svg",       // Invalid dimension
		"/9x1.svg",       // Invalid dimension
		"/1x100.svg",     // Invalid dimension
		"/100x1.svg",     // Invalid dimension
		"/abcx100.svg",   // Missing dimension
		"/100xqwer.svg",  // Missing dimension
		"/500x100.png",   // Invalid extension
		"/500x99999.svg", // Invalid extension

		"/color:00FF00/500x100.svg", // Invalid color value
		"/head:/500x100.svg",        // Missing value
		"/HEAD:default/500x100.svg", // Invalid characters
		"/barf:true/500x100.svg",    // Unrecognized param

	}

	for _, path := range badRequestPaths {
		req, res := fixtures.TestRequest(t, "GET", fmt.Sprintf("http://localhost/avatars%s", path), nil)
		server.router.ServeHTTP(res, req)
		require.Equal(t, http.StatusBadRequest, res.Code)
	}
}

func TestHandleGIFGame_NotFound(t *testing.T) {
	server := NewServer()

	engineServer := fixtures.StubEngineServer(func(res http.ResponseWriter, req *http.Request) {
		res.WriteHeader(http.StatusNotFound)
	})

	req, res := fixtures.TestRequest(t, "GET", "http://localhost/games/GAME_ID/gif", nil)
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

	req, res := fixtures.TestRequest(t, "GET", "http://localhost/games/GAME_ID/gif", nil)
	query := req.URL.Query()
	query.Set("engine_url", engineServer.URL)
	req.URL.RawQuery = query.Encode()

	server.router.ServeHTTP(res, req)

	require.Equal(t, http.StatusOK, res.Code)
	require.Equal(t, "image/gif", res.Result().Header.Get("Content-Type"))
}

func TestHandleGIFFrame_NotFound(t *testing.T) {
	server := NewServer()

	engineServer := fixtures.StubEngineServer(func(res http.ResponseWriter, req *http.Request) {
		res.WriteHeader(http.StatusNotFound)
	})

	req, res := fixtures.TestRequest(t, "GET", "http://localhost/games/GAME_ID/frames/1/gif", nil)
	query := req.URL.Query()
	query.Set("engine_url", engineServer.URL)
	req.URL.RawQuery = query.Encode()

	server.router.ServeHTTP(res, req)
	require.Equal(t, http.StatusNotFound, res.Code)
}

func TestHandleGIFFrame_Success(t *testing.T) {
	fixtures.TestInRootDir()
	server := NewServer()

	engineServer := fixtures.StubEngineServer(func(res http.ResponseWriter, req *http.Request) {
		if strings.HasSuffix(req.URL.Path, "/frames") {
			_, _ = res.Write([]byte(fixtures.ExampleGameFramesResponse))
		} else {
			_, _ = res.Write([]byte(fixtures.ExampleGameResponse))
		}
	})

	req, res := fixtures.TestRequest(t, "GET", "http://localhost/games/GAME_ID/frames/0/gif", nil)
	query := req.URL.Query()
	query.Set("engine_url", engineServer.URL)
	req.URL.RawQuery = query.Encode()

	server.router.ServeHTTP(res, req)

	require.Equal(t, http.StatusOK, res.Code)
	require.Equal(t, "image/gif", res.Result().Header.Get("Content-Type"))
}

func TestHandleASCIIFrame_NotFound(t *testing.T) {
	server := NewServer()

	engineServer := fixtures.StubEngineServer(func(res http.ResponseWriter, req *http.Request) {
		res.WriteHeader(http.StatusNotFound)
	})

	req, res := fixtures.TestRequest(t, "GET", "http://localhost/games/GAME_ID/frames/1/ascii", nil)
	query := req.URL.Query()
	query.Set("engine_url", engineServer.URL)
	req.URL.RawQuery = query.Encode()

	server.router.ServeHTTP(res, req)
	require.Equal(t, http.StatusNotFound, res.Code)
}

func TestHandleASCIIFrame_Success(t *testing.T) {
	fixtures.TestInRootDir()
	server := NewServer()

	engineServer := fixtures.StubEngineServer(func(res http.ResponseWriter, req *http.Request) {
		if strings.HasSuffix(req.URL.Path, "/frames") {
			_, _ = res.Write([]byte(fixtures.ExampleGameFramesResponse))
		} else {
			_, _ = res.Write([]byte(fixtures.ExampleGameResponse))
		}
	})

	req, res := fixtures.TestRequest(t, "GET", "http://localhost/games/GAME_ID/frames/0/ascii", nil)
	query := req.URL.Query()
	query.Set("engine_url", engineServer.URL)
	req.URL.RawQuery = query.Encode()

	server.router.ServeHTTP(res, req)

	require.Equal(t, http.StatusOK, res.Code)
	require.Equal(t, "text/plain; charset=utf-8", res.Result().Header.Get("Content-Type"))
}
