# LyricPiece

A small tool for my command line, gets the currently playing spotify track and then grabs a sample of the lyrics for display. Uses a local boltdb cache for performance.

## Installation

`$ go install github.com/kylegrantlucas/lyricpiece@latest`

## Setup

Using this cli tool requires setting up a Spotify OAuth2 Application so we can call the API for the user's current info.

1. Register an application at: https://developer.spotify.com/my-applications/
  - Use `http://localhost:8080/callback` as the redirect URI

2. Set the `SPOTIFY_ID` environment variable to the client ID you got in step 1.
3. Set the `SPOTIFY_SECRET` environment variable to the client secret from step 1.

## Usage

`$ lyricpiece`