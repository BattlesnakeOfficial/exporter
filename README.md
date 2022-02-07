# Battlesnake Game Exporter

Generate images and animations from Battlesnake games and snakes.

## Local Development

### Install
```
brew install go@1.17
```

Or install manually from https://go.dev/dl/.

### Inkscape

The exporter uses [inkscape](https://inkscape.org/) (version 1.1+) to convert SVGs to PNG format. If Inkscape is not present, the exporter will gracefully handle this by rendering default tails/heads using a local PNG file. You don't have to have inkscape installed to use the exporter unless you want custom heads or tails.

ALSO, some tests which cover functionality that includes SVG rendering will fail unless you have inkscape installed locally.

Inkscape is a freely available, cross-platform tool which you can easily install:

- [Windows](https://inkscape-manuals.readthedocs.io/en/latest/installing-on-windows.html)
- [Mac](https://inkscape-manuals.readthedocs.io/en/latest/installing-on-mac.html)
- [Linux](https://inkscape-manuals.readthedocs.io/en/latest/installing-on-linux.html)

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

Currently there are 3 customizations possible:

- `head` allows you to pick from one of the [available head options](./render/assets/heads)
- `tail` allows you to pick from one of the [available tail options](./render/assets/tails)
- `color` allows you to choose any valid hex code colour. The value must be passed in as a 7 character value like `#cc0033`. Note that the `#` character must be url encoded as `%23`

`curl` example of requesting a single customization

```bash
curl -i http://localhost:8000/avatars/head:beluga/500x100.svg
```

`curl` example of requesting all customization keys

```bash
curl -i http://localhost:8000/avatars/head:beluga/tail:fish/color:%2331688e/500x100.svg
```

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
