# Exporter

Converts a game or frame of a game into a different format.

## Run the server

```shell
PORT=8000 make run
```

## Docker

```shell
docker run -d -p 8000:8000 battlesnakeio/exporter
```

## Endpoints

### `/game/{game id}/frame/{frame offset}?output=raw`

Exports the raw frame from the engine as json.

```shell
curl http://exporter.battlesnake.io/games/15799e31-cd98-4e87-9d49-40ceb4eb439e/frames/30?output=raw
```

### `/game/{game id}/frame/{frame offset}?output=move&youId`

Exports a frame as a move request.

```shell
curl http://exporter.battlesnake.io/games/15799e31-cd98-4e87-9d49-40ceb4eb439e/frames/30?output=move&youId
```

youId = the ID of the snake you want in the `you` field of the move request.  This is a required query parameter.  To find out your snake ID use the `raw` output method above.

### `/game/{game id}/frame/{frame offset}?output=board`

Exports a single frame in one the board format.

Example:

```shell
curl http://exporter.battlesnake.io/games/15799e31-cd98-4e87-9d49-40ceb4eb439e/frames/30?output=board
```

Will output

```none
------------------------
|T1  T2B2B2            |
|B1      B2            |
|B1      H2            |
|B1                    |
|H1                    |
|              T3      |
|        H4B4  B3      |
|        T4B4  B3B3    |
|                H3    |
|          H6B6B6T6    |
|                      |
------------------------
```

H1 - The head of snake 1.  B = Body.  T = Tail.

In this example, snake 5 is dead so it doesn't show on the board.

### `/game/{game id}/frame/{frame offset}?output=board-animated`

Exports the same as board but will reload the page and go to the next frame.

### `/game/{game id}/frame/{frame offset}?output=png`

Exports the game as a png

```shell
curl http://exporter.battlesnake.io/games/15799e31-cd98-4e87-9d49-40ceb4eb439e/frames/30?output=png
```
