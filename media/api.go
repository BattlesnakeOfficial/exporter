package media

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
)

var ErrNotFound = errors.New("resource not found")

func GetAvatarHeadSVG(id string) (string, error) {
	return getAvatarSVG("heads", id)
}

func GetAvatarTailSVG(id string) (string, error) {
	return getAvatarSVG("tails", id)
}

func getAvatarSVG(folder, id string) (string, error) {
	url := fmt.Sprintf("https://media.battlesnake.com/snakes/%s/%s.svg", folder, id)
	client := http.Client{}

	response, err := client.Get(url)
	if err != nil {
		return "", err
	}
	if response.StatusCode == http.StatusNotFound {
		return "", ErrNotFound
	}
	if response.StatusCode != 200 {
		return "", fmt.Errorf("got non 200 from media: %d", response.StatusCode)
	}

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}
