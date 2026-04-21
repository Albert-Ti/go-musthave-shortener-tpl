package model

type ShortenUrlRequest struct {
	Url string `json:"url"`
}

type ShortenUrlResponse struct {
	Result string `json:"result"`
}
