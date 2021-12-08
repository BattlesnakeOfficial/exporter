package http

import (
	"net/http"
	"testing"

	"github.com/BattlesnakeOfficial/exporter/fixtures"
	"github.com/julienschmidt/httprouter"
	"github.com/stretchr/testify/require"
)

func TestHandlesPanic(t *testing.T) {
	server := NewServer()
	server.Router.GET("/fake/panic", httprouter.Handle(func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		panic("an unexpected error")
	}))

	req, res := fixtures.TestRequest(t, "GET", "http://localhost/fake/panic", nil)

	server.Router.ServeHTTP(res, req)
	require.Equal(t, http.StatusInternalServerError, res.Code)
}
