package render_test

import (
	"bytes"
	"encoding/json"
	"image/gif"
	"os"
	"testing"

	"github.com/BattlesnakeOfficial/exporter/engine"
	"github.com/BattlesnakeOfficial/exporter/imagetest"
	"github.com/BattlesnakeOfficial/exporter/render"
	"github.com/stretchr/testify/require"
)

func TestHappyPath(t *testing.T) {

	// Note: there is a known issue with the watermark not loading because the path is wrong when the test executes.

	const goldenFilePath = "testdata/TestHappyPath_golden.gif"
	// Generate the golden file (this shouldn't be done unless tests fail because of intentional rendering changes)
	// generateGoldenFile(t, goldenFilePath) // uncomment to regenerate golden file

	f, err := os.Open(goldenFilePath)
	require.NoError(t, err)
	defer f.Close()
	snapshot, err := gif.Decode(f)
	require.NoError(t, err)

	var buf bytes.Buffer
	game, frame := loadState(t)

	err = render.GameFrameToGIF(&buf, game, frame)
	require.NoError(t, err)
	current, err := gif.Decode(&buf)
	require.NoError(t, err)
	imagetest.Equal(t, snapshot, current)
}

// generates the golden file, uncomment to regenerate
// nolint: unused,deadcode
func generateGoldenFile(t *testing.T, name string) {
	game, frame := loadState(t)
	f, err := os.Create(name)
	require.NoError(t, err)
	defer f.Close()
	err = render.GameFrameToGIF(f, game, frame)
	require.NoError(t, err)
}

func loadState(t *testing.T) (*engine.Game, *engine.GameFrame) {
	var game engine.Game
	err := json.Unmarshal([]byte(gameJSON), &game)
	require.NoError(t, err)
	var frame engine.GameFrame
	err = json.Unmarshal([]byte(frameJSON), &frame)
	require.NoError(t, err)
	return &game, &frame
}

const frameJSON = `{
  "Turn": 150,
  "Food": [
    {
      "X": 7,
      "Y": 1
    },
    {
      "X": 2,
      "Y": 10
    }
  ],
  "Snakes": [
    {
      "ID": "gs_1",
      "Name": "Snake 1",
      "Body": [
        {
          "X": 9,
          "Y": 3
        },
        {
          "X": 9,
          "Y": 4
        },
        {
          "X": 10,
          "Y": 4
        },
        {
          "X": 10,
          "Y": 5
        },
        {
          "X": 10,
          "Y": 6
        },
        {
          "X": 9,
          "Y": 6
        },
        {
          "X": 9,
          "Y": 5
        },
        {
          "X": 8,
          "Y": 5
        },
        {
          "X": 8,
          "Y": 6
        },
        {
          "X": 8,
          "Y": 7
        },
        {
          "X": 7,
          "Y": 7
        },
        {
          "X": 7,
          "Y": 6
        }
      ],
      "Health": 53,
      "Death": null,
      "Color": "#00aacc",
      "HeadType": "crystal-power",
      "TailType": "crystal-power"
    },
    {
      "ID": "gs_2",
      "Name": "snake 2",
      "Body": [
        {
          "X": 6,
          "Y": 6
        },
        {
          "X": 6,
          "Y": 5
        },
        {
          "X": 5,
          "Y": 5
        },
        {
          "X": 5,
          "Y": 4
        },
        {
          "X": 6,
          "Y": 4
        },
        {
          "X": 7,
          "Y": 4
        },
        {
          "X": 8,
          "Y": 4
        },
        {
          "X": 8,
          "Y": 3
        },
        {
          "X": 7,
          "Y": 3
        },
        {
          "X": 6,
          "Y": 3
        },
        {
          "X": 5,
          "Y": 3
        },
        {
          "X": 5,
          "Y": 2
        },
        {
          "X": 6,
          "Y": 2
        },
        {
          "X": 6,
          "Y": 1
        }
      ],
      "Health": 76,
      "Death": null,
      "Color": "#ff7043",
      "HeadType": "trans-rights-scarf",
      "TailType": "rocket"
    },
    {
      "ID": "gs_3",
      "Name": "snake 3",
      "Body": [
        {
          "X": 5,
          "Y": 8
        },
        {
          "X": 5,
          "Y": 7
        },
        {
          "X": 5,
          "Y": 6
        },
        {
          "X": 5,
          "Y": 5
        }
      ],
      "Health": 97,
      "Death": {
        "Cause": "head-collision",
        "Turn": 7
      },
      "Color": "#c91f37",
      "HeadType": "gamer",
      "TailType": "coffee"
    },
    {
      "ID": "gs_4",
      "Name": "Snake 4",
      "Body": [
        {
          "X": 2,
          "Y": 6
        },
        {
          "X": 1,
          "Y": 6
        },
        {
          "X": 0,
          "Y": 6
        },
        {
          "X": 0,
          "Y": 5
        },
        {
          "X": 0,
          "Y": 4
        },
        {
          "X": 1,
          "Y": 4
        },
        {
          "X": 2,
          "Y": 4
        },
        {
          "X": 3,
          "Y": 4
        },
        {
          "X": 3,
          "Y": 5
        },
        {
          "X": 4,
          "Y": 5
        },
        {
          "X": 4,
          "Y": 6
        },
        {
          "X": 4,
          "Y": 7
        }
      ],
      "Health": 54,
      "Death": null,
      "Color": "#ff9900",
      "HeadType": "tiger-king",
      "TailType": "crystal-power"
    }
  ],
  "Hazards": [
    {
      "X": 0,
      "Y": 0
    },
    {
      "X": 0,
      "Y": 1
    },
    {
      "X": 0,
      "Y": 2
    },
    {
      "X": 0,
      "Y": 3
    },
    {
      "X": 0,
      "Y": 4
    },
    {
      "X": 0,
      "Y": 5
    },
    {
      "X": 0,
      "Y": 6
    },
    {
      "X": 0,
      "Y": 7
    },
    {
      "X": 0,
      "Y": 8
    },
    {
      "X": 0,
      "Y": 9
    },
    {
      "X": 0,
      "Y": 10
    },
    {
      "X": 1,
      "Y": 0
    },
    {
      "X": 1,
      "Y": 1
    },
    {
      "X": 1,
      "Y": 2
    },
    {
      "X": 1,
      "Y": 3
    },
    {
      "X": 1,
      "Y": 4
    },
    {
      "X": 1,
      "Y": 5
    },
    {
      "X": 1,
      "Y": 6
    },
    {
      "X": 1,
      "Y": 7
    },
    {
      "X": 1,
      "Y": 8
    },
    {
      "X": 1,
      "Y": 9
    },
    {
      "X": 1,
      "Y": 10
    },
    {
      "X": 2,
      "Y": 0
    },
    {
      "X": 2,
      "Y": 1
    },
    {
      "X": 2,
      "Y": 9
    },
    {
      "X": 2,
      "Y": 10
    },
    {
      "X": 3,
      "Y": 0
    },
    {
      "X": 3,
      "Y": 1
    },
    {
      "X": 3,
      "Y": 9
    },
    {
      "X": 3,
      "Y": 10
    },
    {
      "X": 4,
      "Y": 0
    },
    {
      "X": 4,
      "Y": 1
    },
    {
      "X": 4,
      "Y": 9
    },
    {
      "X": 4,
      "Y": 10
    },
    {
      "X": 5,
      "Y": 0
    },
    {
      "X": 5,
      "Y": 1
    },
    {
      "X": 5,
      "Y": 9
    },
    {
      "X": 5,
      "Y": 10
    },
    {
      "X": 6,
      "Y": 0
    },
    {
      "X": 6,
      "Y": 1
    },
    {
      "X": 6,
      "Y": 9
    },
    {
      "X": 6,
      "Y": 10
    },
    {
      "X": 7,
      "Y": 0
    },
    {
      "X": 7,
      "Y": 1
    },
    {
      "X": 7,
      "Y": 9
    },
    {
      "X": 7,
      "Y": 10
    },
    {
      "X": 8,
      "Y": 0
    },
    {
      "X": 8,
      "Y": 1
    },
    {
      "X": 8,
      "Y": 9
    },
    {
      "X": 8,
      "Y": 10
    },
    {
      "X": 9,
      "Y": 0
    },
    {
      "X": 9,
      "Y": 1
    },
    {
      "X": 9,
      "Y": 9
    },
    {
      "X": 9,
      "Y": 10
    },
    {
      "X": 10,
      "Y": 0
    },
    {
      "X": 10,
      "Y": 1
    },
    {
      "X": 10,
      "Y": 9
    },
    {
      "X": 10,
      "Y": 10
    }
  ]
}`
const gameJSON = `{"ID":"test123","Status":"complete","Width":11,"Height":11}`
