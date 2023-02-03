package media

import (
	"errors"
	"fmt"
	"image"
	"image/color"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/patrickmn/go-cache"
	log "github.com/sirupsen/logrus"
)

var ErrNotFound = errors.New("resource not found")
var mediaServerURL = "https://media.battlesnake.com"

// Create an in-mem media cache (6 hours, evicting every 10 mins)
var mediaCache = cache.New(6*60*time.Minute, 10*time.Minute)

func getCachedMediaResource(path string) (string, error) {
	var resource string

	obj, found := mediaCache.Get(path)
	if found {
		return obj.(string), nil
	}

	resource, err := getMediaResource(path)
	if err != nil {
		return "", err
	}

	mediaCache.Set(path, resource, cache.DefaultExpiration)
	return resource, nil
}

func getMediaResource(path string) (string, error) {
	log.WithField("path", path).Info("fetching media resource")
	url := fmt.Sprintf("%s/%s", mediaServerURL, path)

	client := http.Client{}
	response, err := client.Get(url)
	if err != nil {
		return "", err
	}
	if response.StatusCode == http.StatusNotFound || response.StatusCode == http.StatusForbidden {
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
	return getCachedMediaResource(headSVGPath(id))
}

func GetTailSVG(id string) (string, error) {
	return getCachedMediaResource(tailSVGPath(id))
}

func GetHeadPNG(id string, w, h int, c color.Color) (image.Image, error) {
	return getSnakeSVGImage(headSVGPath(id), fallbackHead, w, h, c)
}

func GetTailPNG(id string, w, h int, c color.Color) (image.Image, error) {
	return getSnakeSVGImage(tailSVGPath(id), fallbackTail, w, h, c)
}

func headSVGPath(id string) string {
	return fmt.Sprintf("snakes/heads/%s.svg", id)
}

func tailSVGPath(id string) string {
	return fmt.Sprintf("snakes/tails/%s.svg", id)
}
