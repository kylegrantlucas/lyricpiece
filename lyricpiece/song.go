package lyricpiece

type Song struct {
	Title  string
	Artist string
}

func (s Song) buildLyricKey() string {
	return s.Artist + "/" + s.Title
}
