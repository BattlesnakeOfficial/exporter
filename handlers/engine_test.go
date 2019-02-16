package handlers

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	gock "gopkg.in/h2non/gock.v1"
)

func TestEngineError(t *testing.T) {
	defer gock.Off()
	GockStatus("status request").ReplyError(fmt.Errorf("Generated error"))
	rr := serveURL("output=raw")
	assert.Equal(t, 500, rr.Code)
	assert.Equal(t, "Problem getting game frames: Get https://engine.battlesnake.io/games/15799e31-cd98-4e87-9d49-40ceb4eb439e/frames?offset=29&limit=1: Generated error", rr.Body.String())
}
func TestEngineErrorResponseCode(t *testing.T) {
	defer gock.Off()
	GockStatus("status request").Response.StatusCode = 503
	rr := serveURL("output=raw")
	assert.Equal(t, 500, rr.Code)
	assert.Equal(t, "Problem getting game frames: Got non 200 response code: 503, message: status request", rr.Body.String())
}
