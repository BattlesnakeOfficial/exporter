# Battlesnake Game Exporter

Converts a game or frame of a game into a different format (mostly images).

## Endpoints

#### `/game/{game id}/gif`

Exports the game as an animated gif.

### `/game/{game id}/frame/{frame number}/gif`

Exports a specific frame as a gif (no animation).

### `/game/{game id}/frame/{frame offset}/ascii`

Exports a specific frame as an ASCII string.
