package http

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
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

func TestHandlerAvatar_OK(t *testing.T) {
	server := NewServer()

	for _, path := range []string{
		"/200x100.svg",
		"/head:beluga/500x100.svg",
		"/head:beluga/tail:fish/color:%2331688e/500x100.svg",
		"/head:beluga/tail:fish/color:%23FfEeCc/500x100.svg",
	} {
		req, res := fixtures.TestRequest(t, "GET", fmt.Sprintf("http://localhost/avatars%s", path), nil)
		server.router.ServeHTTP(res, req)
		require.Equal(t, http.StatusOK, res.Code, path)
	}
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
	defer engineServer.Close()

	req, res := fixtures.TestRequest(t, "GET", "http://localhost/games/GAME_ID/gif", nil)
	query := req.URL.Query()
	query.Set("engine_url", engineServer.URL)
	req.URL.RawQuery = query.Encode()

	server.router.ServeHTTP(res, req)
	require.Equal(t, http.StatusNotFound, res.Code)
}

func TestHandleGIFGame_InvalidResolutions(t *testing.T) {
	fixtures.TestInRootDir()
	server := NewServer()
	req, err := http.NewRequest("GET", "/games/12345678-2666-4a58-9825-1e1cd0c761da/510x510.gif", nil)
	require.NoError(t, err)
	rr := httptest.NewRecorder()

	server.router.ServeHTTP(rr, req)
	require.Equal(t, http.StatusBadRequest, rr.Code)
	require.Equal(t, "Too many pixels! Dimensions 510x510 having resolution 260100 exceeds maximum allowable resolution of 254016.", rr.Body.String())

	req, err = http.NewRequest("GET", "/games/12345678-2666-4a58-9825-1e1cd0c761da/50_50.gif", nil)
	require.NoError(t, err)
	rr = httptest.NewRecorder()

	server.router.ServeHTTP(rr, req)
	require.Equal(t, http.StatusBadRequest, rr.Code)
	require.Equal(t, "Invalid dimensions: \"50_50\" not of the format <WIDTH>x<HEIGHT>.", rr.Body.String())

	engineServer := fixtures.StubEngineServer(func(res http.ResponseWriter, req *http.Request) {
		if strings.HasSuffix(req.URL.Path, "/frames") {
			_, _ = res.Write([]byte(fixtures.ExampleGameFramesResponse))
		} else {
			_, _ = res.Write([]byte(fixtures.ExampleGameResponse))
		}
	})
	defer engineServer.Close()
	req, res := fixtures.TestRequest(t, "GET", "http://localhost/games/GAME_ID/400x400.gif", nil)
	query := req.URL.Query()
	query.Set("engine_url", engineServer.URL)
	req.URL.RawQuery = query.Encode()

	server.router.ServeHTTP(res, req)

	require.Equal(t, http.StatusBadRequest, res.Code)
	require.Equal(t, "Dimensions 400x400 invalid - valid options are: 114x114, 224x224, 334x334, 444x444", res.Body.String())
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
	defer engineServer.Close()

	{
		req, res := fixtures.TestRequest(t, "GET", "http://localhost/games/GAME_ID/gif", nil)
		query := req.URL.Query()
		query.Set("engine_url", engineServer.URL)
		req.URL.RawQuery = query.Encode()

		server.router.ServeHTTP(res, req)

		require.Equal(t, http.StatusOK, res.Code)
		require.Equal(t, "image/gif", res.Result().Header.Get("Content-Type"))
	}

	{
		req, res := fixtures.TestRequest(t, "GET", "http://localhost/games/GAME_ID/444x444.gif", nil)
		query := req.URL.Query()
		query.Set("engine_url", engineServer.URL)
		req.URL.RawQuery = query.Encode()

		server.router.ServeHTTP(res, req)

		require.Equal(t, http.StatusOK, res.Code)
		require.Equal(t, "image/gif", res.Result().Header.Get("Content-Type"))
	}
}

func TestHandleGIFFrame_NotFound(t *testing.T) {
	server := NewServer()

	engineServer := fixtures.StubEngineServer(func(res http.ResponseWriter, req *http.Request) {
		res.WriteHeader(http.StatusNotFound)
	})
	defer engineServer.Close()

	req, res := fixtures.TestRequest(t, "GET", "http://localhost/games/GAME_ID/frames/1/gif", nil)
	query := req.URL.Query()
	query.Set("engine_url", engineServer.URL)
	req.URL.RawQuery = query.Encode()

	server.router.ServeHTTP(res, req)
	require.Equal(t, http.StatusNotFound, res.Code)
}

func TestHandleGIFFrame_InvalidResolutions(t *testing.T) {
	fixtures.TestInRootDir()
	server := NewServer()
	req, err := http.NewRequest("GET", "/games/12345678-2666-4a58-9825-1e1cd0c761da/frames/1/510x510.gif", nil)
	require.NoError(t, err)
	rr := httptest.NewRecorder()

	server.router.ServeHTTP(rr, req)
	require.Equal(t, http.StatusBadRequest, rr.Code)
	require.Equal(t, "Too many pixels! Dimensions 510x510 having resolution 260100 exceeds maximum allowable resolution of 254016.", rr.Body.String())

	req, err = http.NewRequest("GET", "/games/12345678-2666-4a58-9825-1e1cd0c761da/frames/1/50_50.gif", nil)
	require.NoError(t, err)
	rr = httptest.NewRecorder()

	server.router.ServeHTTP(rr, req)
	require.Equal(t, http.StatusBadRequest, rr.Code)
	require.Equal(t, "Invalid dimensions: \"50_50\" not of the format <WIDTH>x<HEIGHT>.", rr.Body.String())

	engineServer := fixtures.StubEngineServer(func(res http.ResponseWriter, req *http.Request) {
		if strings.HasSuffix(req.URL.Path, "/frames") {
			_, _ = res.Write([]byte(fixtures.ExampleGameFramesResponse))
		} else {
			_, _ = res.Write([]byte(fixtures.ExampleGameResponse))
		}
	})
	defer engineServer.Close()
	req, res := fixtures.TestRequest(t, "GET", "http://localhost/games/GAME_ID/frames/1/400x400.gif", nil)
	query := req.URL.Query()
	query.Set("engine_url", engineServer.URL)
	req.URL.RawQuery = query.Encode()

	server.router.ServeHTTP(res, req)

	require.Equal(t, http.StatusBadRequest, res.Code)
	require.Equal(t, "Dimensions 400x400 invalid - valid options are: 114x114, 224x224, 334x334, 444x444", res.Body.String())
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
	defer engineServer.Close()

	{
		req, res := fixtures.TestRequest(t, "GET", "http://localhost/games/GAME_ID/frames/0/gif", nil)
		query := req.URL.Query()
		query.Set("engine_url", engineServer.URL)
		req.URL.RawQuery = query.Encode()

		server.router.ServeHTTP(res, req)

		require.Equal(t, http.StatusOK, res.Code)
		require.Equal(t, "image/gif", res.Result().Header.Get("Content-Type"))
	}

	{
		req, res := fixtures.TestRequest(t, "GET", "http://localhost/games/GAME_ID/frames/0/334x334.gif", nil)
		query := req.URL.Query()
		query.Set("engine_url", engineServer.URL)
		req.URL.RawQuery = query.Encode()

		server.router.ServeHTTP(res, req)

		require.Equal(t, http.StatusOK, res.Code)
		require.Equal(t, "image/gif", res.Result().Header.Get("Content-Type"))
	}
}

func TestHandleASCIIFrame_NotFound(t *testing.T) {
	server := NewServer()

	engineServer := fixtures.StubEngineServer(func(res http.ResponseWriter, req *http.Request) {
		res.WriteHeader(http.StatusNotFound)
	})
	defer engineServer.Close()

	{
		req, res := fixtures.TestRequest(t, "GET", "http://localhost/games/GAME_ID/frames/1/ascii", nil)
		query := req.URL.Query()
		query.Set("engine_url", engineServer.URL)
		req.URL.RawQuery = query.Encode()

		server.router.ServeHTTP(res, req)
		require.Equal(t, http.StatusNotFound, res.Code)
	}

	{
		req, res := fixtures.TestRequest(t, "GET", "http://localhost/games/GAME_ID/frames/1.txt", nil)
		query := req.URL.Query()
		query.Set("engine_url", engineServer.URL)
		req.URL.RawQuery = query.Encode()

		server.router.ServeHTTP(res, req)
		require.Equal(t, http.StatusNotFound, res.Code)
	}
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
	defer engineServer.Close()

	{
		req, res := fixtures.TestRequest(t, "GET", "http://localhost/games/GAME_ID/frames/0/ascii", nil)
		query := req.URL.Query()
		query.Set("engine_url", engineServer.URL)
		req.URL.RawQuery = query.Encode()

		server.router.ServeHTTP(res, req)

		require.Equal(t, http.StatusOK, res.Code)
		require.Equal(t, "text/plain; charset=utf-8", res.Result().Header.Get("Content-Type"))
	}

	{
		req, res := fixtures.TestRequest(t, "GET", "http://localhost/games/GAME_ID/frames/0.txt", nil)
		query := req.URL.Query()
		query.Set("engine_url", engineServer.URL)
		req.URL.RawQuery = query.Encode()

		server.router.ServeHTTP(res, req)

		require.Equal(t, http.StatusOK, res.Code)
		require.Equal(t, "text/plain; charset=utf-8", res.Result().Header.Get("Content-Type"))
	}
}

func TestValidateGIFSize(t *testing.T) {

	require.NoError(t, validateGIFSize(0, 0), "0 values should default to a calculated GIF size")

	// test some allowable sizes
	require.NoError(t, validateGIFSize(10, 10))
	require.NoError(t, validateGIFSize(114, 114))
	require.NoError(t, validateGIFSize(384, 114))
	require.NoError(t, validateGIFSize(504, 504))

	// test some non-allowable sizes
	require.Equal(t, errors.New("Invalid width -1: cannot be < 0."), validateGIFSize(-1, 100))
	require.Equal(t, errors.New("Invalid height -1: cannot be < 0"), validateGIFSize(100, -1))

	// test too high resolutions
	require.Equal(t, errors.New("Too many pixels! Dimensions 505x505 having resolution 255025 exceeds maximum allowable resolution of 254016."), validateGIFSize(505, 505))
	require.Error(t, validateGIFSize(384, 1995))
	require.Error(t, validateGIFSize(1995, 384))
}
