package lyricpiece

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/boltdb/bolt"
	lyrics "github.com/rhnvrm/lyric-api-go"
)

type Client struct {
	db *bolt.DB
}

func NewClient(dbfile string) (*Client, error) {
	db, err := bolt.Open(dbfile, 0600, nil)
	if err != nil {
		return nil, err
	}

	return &Client{
		db: db,
	}, nil
}

func (lp *Client) Close() error {
	return lp.db.Close()
}

func (lp *Client) GetLyricPiece(song Song) (string, error) {
	rand.Seed(time.Now().Unix())

	lyrics, err := lp.getLyrics(lp.db, song)
	if err != nil {
		fmt.Println(err)
		return "", err
	}
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
		), nil
	}

	return "", nil
}

func (lp *Client) getLyrics(db *bolt.DB, song Song) (string, error) {
	lyrics, err := lp.queryDBForLyrics(db, song)
	if err != nil {
		return lyrics, err
	}

	if lyrics == "" {
		downloadedLyrics, err := lp.queryLyrics(song)
		if err != nil {
			return lyrics, err
		}

		if downloadedLyrics != "" {
			err = lp.writeLyricsToDB(db, song, downloadedLyrics)
			if err != nil {
				return lyrics, err
			}
		}

		lyrics = downloadedLyrics
	}

	return lyrics, nil
}

func (lp *Client) queryDBForLyrics(db *bolt.DB, song Song) (string, error) {
	var lyrics string
	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("lyrics"))
		if b == nil {
			// we don't care if we can't find the bucket
			return nil
		}
		lyrics = string(b.Get([]byte(song.buildLyricKey())))

		return nil
	})
	if err != nil {
		return lyrics, err
	}

	return lyrics, nil
}

func (lp *Client) writeLyricsToDB(db *bolt.DB, song Song, lyrics string) error {
	err := db.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte("lyrics"))
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}
		return bucket.Put([]byte(song.buildLyricKey()), []byte(lyrics))
	})
	if err != nil {
		return err
	}

	return nil
}

func (lp *Client) queryLyrics(song Song) (string, error) {
	l := lyrics.New()
	lyric, err := l.Search(song.Artist, song.Title)
	if err != nil {
		return "", err
	}

	return lyric, nil
}
