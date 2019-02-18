package handlers

import (
	"testing"
)

func TestBox(t *testing.T) {
	_, err := GetSnakeHeadImage("beluga")
	if err != nil {
		panic(err)
	}
}
