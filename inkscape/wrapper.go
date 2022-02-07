package inkscape

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"image/png"
	"os"
	"os/exec"
)

var defaultCommand = "inkscape"

// Client is an implementation of an Inkscape CLI wrapper.
// It assumes a locally available inkscape command can be run.
type Client struct {
	// the name/path for the inkscape CLI command
	// If left empty, the default inkscape command will be used.
	Command string
}

// IsAvailable checks whether the inkscape command is locally available.
// Provides feature detection for a sort of progressive enhancement.
func (c Client) IsAvailable() bool {
	_, err := exec.LookPath(c.cmd())
	return err == nil
}

// SVGToPNG raserizes the SVG at the specified path to PNG format.
func (c Client) SVGToPNG(path string, width, height int) (image.Image, error) {
	if height < 1 {
		return nil, errors.New("invalid height")
	}
	if width < 1 {
		return nil, errors.New("invalid width")
	}

	_, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	cmd := exec.Command(c.cmd(), path, "-w", fmt.Sprint(width), "-h", fmt.Sprint(height), "--export-type=png", "--export-filename=-")
	b := bytes.NewBuffer(nil)
	cmd.Stdout = b
	err = cmd.Start()
	if err != nil {
		return nil, err
	}
	err = cmd.Wait()
	if err != nil {
		return nil, err
	}

	// if we get no bytes on stdout, that means something went wrong
	if b.Len() == 0 {
		return nil, errors.New("error processing SVG")
	}

	img, err := png.Decode(b)
	return img, err
}

func (c Client) cmd() string {
	if c.Command == "" {
		return defaultCommand
	}
	return c.Command
}
