package repository

import "strconv"

type ShortenedUrls struct {
	List  map[string]string
	Count uint
}

func (u *ShortenedUrls) GetUrl(key string) string {
	return u.List[key]
}

func (u *ShortenedUrls) Set(url string) string {
	key := "key_" + strconv.Itoa(int(u.Count))
	u.List[key] = url
	u.Count++

	return key
}
