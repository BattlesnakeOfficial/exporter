package http

import (
	"net/http"
	"testing"

	"github.com/BattlesnakeOfficial/exporter/fixtures"
	"github.com/stretchr/testify/require"
	"goji.io/pat"
)

func TestHandlesPanic(t *testing.T) {
	server := NewServer()
	server.router.HandleFunc(pat.Get("/fake/panic"), func(w http.ResponseWriter, r *http.Request) {
		panic("an unexpected error")
	})

	req, res := fixtures.TestRequest(t, "GET", "http://localhost/fake/panic", nil)

	server.router.ServeHTTP(res, req)
	require.Equal(t, http.StatusInternalServerError, res.Code)
}
