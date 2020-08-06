package main

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/andybrewer/mack"
	tilde "gopkg.in/mattes/go-expand-tilde.v1"
)

func main() {
	song := getCurrentSong()

	lyrics, err := getLyrics(song)
	if err != nil {
		panic(err)
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
	req, err := http.NewRequest("GET", "https://makeitpersonal.co/lyrics", nil)
	if err != nil {
		return "", err
	}

	q := req.URL.Query()
	q.Add("artist", song.Artist)
	q.Add("title", song.Title)
	req.URL.RawQuery = q.Encode()

	timeout := time.Duration(5 * time.Second)
	client := http.Client{
		Timeout: timeout,
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if strings.TrimSpace(string(body)) == "Sorry, We don't have lyrics for this song yet. Add them to https://lyrics.wikia.com" {
		return "", nil
	}

	return string(body), nil
}
