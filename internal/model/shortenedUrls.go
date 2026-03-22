package model

import "strconv"

type ShortenedUrls struct {
	List  map[uint]string
	Count uint
}

func (u *ShortenedUrls) GetUrl(key string) string {
	for k, v := range u.List {
		if "key_"+strconv.Itoa(int(k)) == key {
			return v
		}
	}
	return ""
}
func (u *ShortenedUrls) Set(url string) string {
	u.List[u.Count] = url

	k := u.Count
	u.Count++
	return "key_" + strconv.Itoa(int(k))
}
