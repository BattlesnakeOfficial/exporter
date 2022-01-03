# Battlesnake Game Exporter

Generate images and animations from Battlesnake games and snakes.

## Local Development

### Install
```
brew install go@1.17
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

#### `/avatars/{customation_key:customization_value}/{width}x{height}.svg`

Exports a Battlesnake avatar with the provided customizations.

#### `/games/{game id}/gif`

Exports the game as an animated gif.

#### `/games/{game id}/frames/{frame number}/gif`

Exports a specific frame as a gif (no animation).

#### `/games/{game id}/frames/{frame number}/ascii`

Exports a specific frame as an ASCII string.

## Caching

By default all exported objects are set to be cached by the browser for 24 hours.

## Feedback

* **Do you have an issue or suggestions for this repository?** Head over to our [Feedback Repository](https://play.battlesnake.com/feedback) today and let us know!
