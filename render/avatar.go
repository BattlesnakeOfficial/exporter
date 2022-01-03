package render

import (
	"bytes"
	"errors"
	"regexp"
	"text/template"
)

var ErrInvalidAvatarSettings = errors.New("invalid avatar settings")

type AvatarSettings struct {
	Height  int
	Width   int
	HeadSVG string
	TailSVG string
	Color   string
}

func (a AvatarSettings) CalculateBodyWidth() int {
	return a.Width - (2 * a.Height)
}

func (a AvatarSettings) CalculateHeadOffset() int {
	return a.Width - a.Height
}

func (a AvatarSettings) Validate() bool {
	// Nothing larger than 10000x10000
	if a.Height > 10000 || a.Width > 10000 {
		return false
	}
	// Body length must be non-negative
	if a.CalculateBodyWidth() < 0 {
		return false
	}
	return true
}

const avatarTemplate = `<svg id="root" xmlns="http://www.w3.org/2000/svg" fill="{{ .Color }}" width="{{ .Width }}" height="{{ .Height }}">
<g transform="scale(-1, 1) translate(-{{ .Height }}, 0)">
	<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 100 100" width="{{ .Height }}" height="{{ .Height }}">
		{{ .TailSVG }}
	</svg>
</g>
<g transform="translate({{ .Height }}, 0)">
	<rect width="{{ .CalculateBodyWidth }}" height="{{ .Height }}" />
</g>
<g transform="translate({{ .CalculateHeadOffset }}, 0)">
	<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 100 100" width="{{ .Height }}" height="{{ .Height }}">
		{{ .HeadSVG }}
	</svg>
</g>
</svg>`

func AvatarSVG(settings AvatarSettings) (string, error) {
	if !settings.Validate() {
		return "", ErrInvalidAvatarSettings
	}

	t := template.Must(template.New("avatar").Parse(avatarTemplate))

	re := regexp.MustCompile(`^<svg ([^>]*)>`)
	settings.HeadSVG = re.ReplaceAllString(settings.HeadSVG, "")
	settings.TailSVG = re.ReplaceAllString(settings.TailSVG, "")

	re = regexp.MustCompile(`</svg>\s*$`)
	settings.HeadSVG = re.ReplaceAllString(settings.HeadSVG, "")
	settings.TailSVG = re.ReplaceAllString(settings.TailSVG, "")

	buf := &bytes.Buffer{}
	_ = t.Execute(buf, settings)
	return buf.String(), nil
}
