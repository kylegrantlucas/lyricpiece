package main

import (
	"fmt"
	"os"

	"github.com/kylegrantlucas/lyricpiece/lyricpiece"
	"github.com/kylegrantlucas/lyricpiece/spotify"
)

func main() {
	dbPath, err := buildDBPath()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}

	// call to spotify to get the currently playing song
	spotClient, err := spotify.NewClient(dbPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}

	song, err := spotClient.GetCurrentlyPlaying()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}

	// close up the DB so the lyricpiece client can use it
	err = spotClient.Close()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}

	client, err := lyricpiece.NewClient(dbPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
	defer client.Close()

	if song != nil {
		lyricPiece, err := client.GetLyricPiece(*song)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
		}

		fmt.Print(lyricPiece)
	}
}

func buildDBPath() (string, error) {
	homedir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	path := homedir + "/.cache/lyricpiece"
	err = os.MkdirAll(path, os.ModePerm)
	if err != nil {
		return "", err
	}

	return path + "/lyricpiece.db", nil
}
