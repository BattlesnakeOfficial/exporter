package media

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/sirupsen/logrus"
)

var ErrNotFound = errors.New("resource not found")

func getMediaResource(path string) (string, error) {
	logrus.WithField("path", path).Info("fetching media resource")
	url := fmt.Sprintf("https://media.battlesnake.com/%s", path)

	client := http.Client{}
	response, err := client.Get(url)
	if err != nil {
		return "", err
	}
	if response.StatusCode == http.StatusNotFound {
		return "", ErrNotFound
	}
	if response.StatusCode != 200 {
		return "", fmt.Errorf("got non 200 from media '%s': %d", path, response.StatusCode)
	}

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

func GetHeadSVG(id string) (string, error) {
	return getCachedMediaResource(fmt.Sprintf("snakes/heads/%s.svg", id))
}

func GetTailSVG(id string) (string, error) {
	return getCachedMediaResource(fmt.Sprintf("snakes/tails/%s.svg", id))
}
