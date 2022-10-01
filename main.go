package main

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/andybrewer/mack"
	lyrics "github.com/rhnvrm/lyric-api-go"
	tilde "gopkg.in/mattes/go-expand-tilde.v1"
)

func main() {
	song := getCurrentSong()

	lyrics, err := getLyrics(song)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}

	lyricPiece := getRandomLyricPiece(lyrics)

	fmt.Print(lyricPiece)
}

type Song struct {
	Title  string
	Artist string
}

func getCurrentSong() Song {
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

	return Song{Title: track, Artist: artist}
}

func getLyrics(song Song) (string, error) {
	var lyrics string
	filename, err := tilde.Expand(fmt.Sprintf("~/lyrics/%s-%s.txt", strings.TrimSpace(song.Title), strings.TrimSpace(song.Artist)))
	if err != nil {
		return lyrics, err
	}

	if fileExists(filename) {
		file, err := ioutil.ReadFile(filename)
		if err != nil {
			return lyrics, err
		}
		lyrics = string(file)
	} else {
		downloadedLyrics, err := queryLyrics(song)
		if err != nil {
			return lyrics, err
		}

		if downloadedLyrics != "" {
			err = ioutil.WriteFile(filename, []byte(downloadedLyrics), 0644)
			if err != nil {
				return lyrics, err
			}
		}

		lyrics = downloadedLyrics
	}

	return lyrics, nil
}

func getRandomLyricPiece(lyrics string) string {
	rand.Seed(time.Now().Unix())

	chunkedLyrics := strings.Split(lyrics, "\n\n")

	fullLyrics := []string{}
	for _, s := range chunkedLyrics {
		if strings.TrimSpace(s) != "" {
			fullLyrics = append(fullLyrics, s)
		}
	}

	if len(fullLyrics) > 0 {
		return strings.TrimSpace(
			fullLyrics[rand.Intn(len(fullLyrics))],
		)
	}

	return ""
}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func queryLyrics(song Song) (string, error) {
	l := lyrics.New()
	lyric, err := l.Search(song.Artist, song.Title)
	if err != nil {
		return "", err
	}

	return lyric, nil
}
