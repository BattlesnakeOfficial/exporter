package fixtures

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"runtime"
	"testing"

	"github.com/stretchr/testify/require"
)

// Helper for setting up an engine endpoint that returns a stub response
func StubEngineServer(handler http.HandlerFunc) *httptest.Server {
	return httptest.NewServer(handler)
}

// Helper for creating a request and recording the response
func TestRequest(t *testing.T, method, url string, body io.Reader) (*http.Request, *httptest.ResponseRecorder) {
	clientRequest, err := http.NewRequest(method, url, body)
	require.NoError(t, err)

	recorder := httptest.NewRecorder()

	return clientRequest, recorder
}

// Helper to change the current directory to the project root, instead of the directory the test is located in.
// Required in some tests to ensure assets are located correctly.
func TestInRootDir() {
	_, filename, _, _ := runtime.Caller(0)
	root := path.Join(path.Dir(filename), "..")
	_ = os.Chdir(root)
}

const ExampleGameResponse = `
{
    "Game": {
        "ID": "1f26bc6e-2b96-4f54-ba01-f610e68e2d45",
        "Status": "complete",
        "Width": 11,
        "Height": 11,
        "Ruleset": {
            "foodSpawnChance": "15",
            "minimumFood": "1",
            "name": "standard"
        },
        "SnakeTimeout": 500,
        "MaxTurns": 0,
        "FoodSpawns": [],
        "HazardSpawns": [],
        "Source": "arena"
    },
    "LastFrame": {
        "Turn": 1,
        "Snakes": [
            {
                "ID": "snake1",
                "Name": "Snake",
                "URL": "",
                "Body": [
                    {
                        "X": 10,
                        "Y": 11
                    },
                    {
                        "X": 10,
                        "Y": 10
                    },
                    {
                        "X": 10,
                        "Y": 10
                    }
                ],
                "Health": 99,
                "Death": {
                    "Cause":"wall-collision",
                    "Turn":1,
                    "EliminatedBy":""
                },
                "Color": "#123456",
                "HeadType": "silly",
                "TailType": "bolt",
                "Latency": "185",
                "Shout": "",
                "Squad": "",
                "APIVersion": "",
                "Author": "",
                "StatusCode": 200,
                "Error": "",
                "TimingMicros": {},
                "IsBot": false,
                "IsEnvironment": false,
                "ProxyURL": ""
            }
        ],
        "Food": [
            {
                "X": 8,
                "Y": 6
            }
        ],
        "Hazards": [
            {
                "X": 4,
                "Y": 2
            }
        ]
    }
}
`

const ExampleGameFramesResponse = `
{
    "Count": 2,
    "Frames": [
        {
            "Turn": 0,
            "Snakes": [
                {
                    "ID": "snake1",
                    "Name": "Snake",
                    "URL": "",
                    "Body": [
                        {
                            "X": 10,
                            "Y": 10
                        },
                        {
                            "X": 10,
                            "Y": 10
                        },
                        {
                            "X": 10,
                            "Y": 10
                        }
                    ],
                    "Health": 100,
                    "Death": null,
                    "Color": "#123456",
                    "HeadType": "silly",
                    "TailType": "bolt",
                    "Latency": "185",
                    "Shout": "",
                    "Squad": "",
                    "APIVersion": "",
                    "Author": "",
                    "StatusCode": 200,
                    "Error": "",
                    "TimingMicros": {},
                    "IsBot": false,
                    "IsEnvironment": false,
                    "ProxyURL": ""
                }
            ],
            "Food": [
                {
                    "X": 8,
                    "Y": 6
                }
            ],
            "Hazards": [
                {
                    "X": 4,
                    "Y": 2
                }
            ]
        },
        {
            "Turn": 1,
            "Snakes": [
                {
                    "ID": "snake1",
                    "Name": "Snake",
                    "URL": "",
                    "Body": [
                        {
                            "X": 10,
                            "Y": 11
                        },
                        {
                            "X": 10,
                            "Y": 10
                        },
                        {
                            "X": 10,
                            "Y": 10
                        }
                    ],
                    "Health": 99,
                    "Death": {
                        "Cause":"wall-collision",
                        "Turn":1,
                        "EliminatedBy":""
                    },
                    "Color": "#123456",
                    "HeadType": "silly",
                    "TailType": "bolt",
                    "Latency": "185",
                    "Shout": "",
                    "Squad": "",
                    "APIVersion": "",
                    "Author": "",
                    "StatusCode": 200,
                    "Error": "",
                    "TimingMicros": {},
                    "IsBot": false,
                    "IsEnvironment": false,
                    "ProxyURL": ""
                }
            ],
            "Food": [
                {
                    "X": 8,
                    "Y": 6
                }
            ],
            "Hazards": [
                {
                    "X": 4,
                    "Y": 2
                }
            ]
        }
    ]
}
`
