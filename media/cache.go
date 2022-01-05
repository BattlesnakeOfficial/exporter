package media

import (
	"time"

	"github.com/patrickmn/go-cache"
)

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
