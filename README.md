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

#### local

```sh
go build -o bin/exporter cmd/exporter/main.go
```

#### docker

```sh
docker build . -t exporter
```

### Running the server


#### local

```
export PORT=8000 # optional port override

./bin/exporter
```

#### docker

```sh
docker run -it -p 8000:8000 exporter:latest
```

### Running the tests
```
go test ./...
```

**Note:** some tests may fail if `inkscape` is not available **or** if you have the wrong version of `inkscape` installed (pre `1.x` versions of `inkscape` won't work)

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

**DEPRECATED:** use `/games/{game id}/{width}x{height}.gif` instead.

Exports the game as an animated gif.

#### `/games/{game id}/frames/{frame number}/gif`

**DEPRECATED:** use `/games/{game id}/frames/{frame number}/{width}x{height}.gif` instead.

Exports a specific frame as a gif (no animation).

#### `/games/{game id}/frames/{frame number}/ascii`

**DEPRECATED:** use `/games/{game id}/frames/{frame number}.txt` instead.

Exports a specific frame as an ASCII string.

#### `/games/{game id}/{width}x{height}.gif`

Exports the game as an animated gif sized `width` pixels wide and `height` pixels high.

See [GIF size validation](#Choose-a-GIF-size) for details about how to choose a valid GIF resolution

#### `/games/{game id}/frames/{frame number}/{width}x{height}.gif`

Exports the game as an animated gif sized `width` pixels wide and `height` pixels high.

See [GIF size validation](#Choose-a-GIF-size) for details about how to choose a valid GIF resolution

#### `/games/{game id}/frames/{frame number}.txt`

Exports a specific frame as an ASCII string.

### Choose a GIF size

GIF sizes are restricted to a limited set of options based on the game board being exported. Additionally, there is an upper-limit of a maximum resolution of `504x504` (`254016` pixels) which supersedes the calculation of available options.


At the time of this writing, the options are:

- 10 pixels per board square (+ 4 pixels for border)
- 20 pixels per board square (+ 4 pixels for border)
- 30 pixels per board square (+ 4 pixels for border)
- 40 pixels per board square (+ 4 pixels for border)

These options may change. If you need an up-to-date list, look at `allowedPixelsPerSquare` in [handlers.go](./http/handlers.go) for an authoritative list of allowed resolutions.

Using the above, you can determine the available GIF dimensions that you can request.

Examples:
- allowed sizes allowed for **11x11** board are: 
  - **114x114** (10 pixels per board square)
  - **224x224** (20 pixels per board square)
  - **334x334** (30 pixels per board square)
  - **444x444** (40 pixels per board square)
- allowed sizes allowed for **19x19** board are: 
  - **194x194** (10 pixels per board square)
  - **384x384** (20 pixels per board square)
  - **574x574** (30 pixels per board square) (**Disallowed** because it exceeds `504x504`)
  - **764x764** (40 pixels per board square) (**Disallowed** because it exceeds `504x504`)
- etc...

## Caching

By default all exported objects are set to be cached by the browser for 24 hours.

## Feedback

* **Do you have an issue or suggestions for this repository?** Head over to our [Feedback Repository](https://play.battlesnake.com/feedback) today and let us know!
