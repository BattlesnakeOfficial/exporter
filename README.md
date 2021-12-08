# Battlesnake Game Exporter

Converts a game or frame of a game into a different format (mostly images).

## Local Development

### Install
```
brew install go@1.15
```

Or install manually from https://go.dev/dl/.

### Building
Installing to `$GOPATH/bin/`: (defaults to `$HOME/go/bin` if you don't have GOPATH set):
```
go install -v ./cmd/exporter
```

### Running the server
```
export PORT=8000 # optional port override

exporter
```

### Running the tests
```
go test ./...
```

## Endpoints

#### `/games/{game id}/gif`

Exports the game as an animated gif.

#### `/games/{game id}/frames/{frame number}/gif`

Exports a specific frame as a gif (no animation).

#### `/games/{game id}/frames/{frame number}/ascii`

Exports a specific frame as an ASCII string.

### Feedback

* **Do you have an issue or suggestions for this repository?** Head over to our [Feedback Repository](https://play.battlesnake.com/feedback) today and let us know!
