package main

import (
	"fmt"
	"os"
	"sync"

	"github.com/andybrewer/mack"
	"github.com/kylegrantlucas/lyricpiece/lyricpiece"
)

func main() {
	dbPath, err := buildDBPath()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}

	client, err := lyricpiece.NewClient(dbPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
	defer client.Close()

	song := getCurrentSong()
	lyricPiece, err := client.GetLyricPiece(song)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}

	fmt.Print(lyricPiece)
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

func getCurrentSong() lyricpiece.Song {
	var track, artist string
	wg := &sync.WaitGroup{}
	wg.Add(2)

	go func() {
		track, _ = mack.Tell("Spotify", "name of current track as string")
		wg.Done()
	}()

	go func() {
		artist, _ = mack.Tell("Spotify", "artist of current track as string")
		wg.Done()
	}()

	wg.Wait()

	return lyricpiece.Song{Title: track, Artist: artist}
}
