# Battlesnake Game Exporter

Converts a game or frame of a game into a different format (mostly images).

## Endpoints

#### `/games/{game id}/gif`

Exports the game as an animated gif.

#### `/games/{game id}/frames/{frame number}/gif`

Exports a specific frame as a gif (no animation).

#### `/games/{game id}/frames/{frame number}/ascii`

Exports a specific frame as an ASCII string.
