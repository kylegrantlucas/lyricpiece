// 1. Register an application at: https://developer.spotify.com/my-applications/
//   - Use "http://localhost:8080/callback" as the redirect URI

// 2. Set the SPOTIFY_ID environment variable to the client ID you got in step 1.
// 3. Set the SPOTIFY_SECRET environment variable to the client secret from step 1.
package spotify

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/boltdb/bolt"
	"github.com/kylegrantlucas/lyricpiece/lyricpiece"
	spot "github.com/zmb3/spotify/v2"
	spotifyauth "github.com/zmb3/spotify/v2/auth"
	"golang.org/x/oauth2"
)

const redirectURI = "http://localhost:8080/callback"

var (
	auth  = spotifyauth.New(spotifyauth.WithRedirectURL(redirectURI), spotifyauth.WithScopes(spotifyauth.ScopeUserReadCurrentlyPlaying, spotifyauth.ScopeUserReadPlaybackState))
	ch    = make(chan *oauth2.Token)
	state = "abc123"
)

type Client struct {
	db     *bolt.DB
	client *spot.Client
}

func NewClient(dbPath string) (*Client, error) {
	db, err := bolt.Open(dbPath, 0600, nil)
	if err != nil {
		return nil, err
	}

	client := &Client{
		db: db,
	}

	client.auth()

	return client, nil
}

func (c *Client) Close() error {
	return c.db.Close()
}

// get currently playing song
func (c *Client) GetCurrentlyPlaying() (*lyricpiece.Song, error) {
	currentlyPlaying, err := c.client.PlayerCurrentlyPlaying(context.Background())
	if err != nil {
		return nil, err
	}

	if currentlyPlaying.Playing {
		return &lyricpiece.Song{
			Title:  currentlyPlaying.Item.Name,
			Artist: currentlyPlaying.Item.Artists[0].Name,
		}, nil
	}

	return nil, nil
}

func (c *Client) queryDBForToken() (*oauth2.Token, error) {
	var token *oauth2.Token
	err := c.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte("spotify"))
		if bucket == nil {
			return nil
		}

		tokenJSON := bucket.Get([]byte("token"))
		if tokenJSON == nil {
			return nil
		}

		return json.Unmarshal(tokenJSON, &token)
	})
	if err != nil {
		return nil, err
	}

	return token, nil
}

func (c *Client) auth() error {
	// try to get the token from the database
	token, err := c.queryDBForToken()
	if err != nil {
		return err
	}

	// if we don't have a token, get one
	if token == nil {
		url := auth.AuthURL(state)
		log.Println("Please log in to Spotify by visiting the following page in your browser:", url)

		// first start an HTTP server
		http.HandleFunc("/callback", completeAuth)
		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			log.Println("Got request for:", r.URL.String())
		})
		go func() {
			err := http.ListenAndServe(":8080", nil)
			if err != nil {
				log.Fatal(err)
			}
		}()

		// wait for auth to complete
		token = <-ch

		// store the token in bolt
		err := c.db.Update(func(tx *bolt.Tx) error {
			bucket, err := tx.CreateBucketIfNotExists([]byte("spotify"))
			if err != nil {
				return fmt.Errorf("create bucket: %s", err)
			}

			// marshal the token to json
			tokenJSON, err := json.Marshal(token)
			if err != nil {
				return fmt.Errorf("marshal token: %s", err)
			}

			err = bucket.Put([]byte("token"), []byte(tokenJSON))
			if err != nil {
				return fmt.Errorf("put token: %s", err)
			}

			return nil
		})
		if err != nil {
			return err
		}
	}

	// use the token to get an authenticated client
	client := spot.New(auth.Client(context.Background(), token))
	c.client = client

	return nil
}

func completeAuth(w http.ResponseWriter, r *http.Request) {
	tok, err := auth.Token(r.Context(), state, r)
	if err != nil {
		http.Error(w, "Couldn't get token", http.StatusForbidden)
		log.Fatal(err)
	}
	if st := r.FormValue("state"); st != state {
		http.NotFound(w, r)
		log.Fatalf("State mismatch: %s != %s\n", st, state)
	}

	ch <- tok
}
