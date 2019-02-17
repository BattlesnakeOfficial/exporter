package handlers

import (
	"fmt"
	"testing"
)

func TestBox(t *testing.T) {
	image, err := GetSnakeHeadImage("beluga")
	if err != nil {
		panic(err)
	}
	fmt.Println(image)
}
