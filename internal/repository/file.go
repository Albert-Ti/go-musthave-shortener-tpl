package repository

import (
	"encoding/json"
	"io"
	"os"
)

type FileStorage struct {
	element *os.File
	encoder *json.Encoder
	urls    map[string]string
}

type fileUrlRecord struct {
	Uuid        int    `json:"uuid"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

func NewFileStorage(path string) (*FileStorage, error) {
	file, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}

	urls := map[string]string{}
	decoder := json.NewDecoder(file)

	for {
		var record fileUrlRecord
		err := decoder.Decode(&record)
		if err == io.EOF {
			break
		}
		if err != nil {
			panic(err)
		}
		urls[record.ShortURL] = record.OriginalURL
	}

	return &FileStorage{
		urls:    urls,
		element: file,
		encoder: json.NewEncoder(file),
	}, nil
}

func (u *FileStorage) Get(key string) string {
	return u.urls[key]
}

func (u *FileStorage) Save(key string, url string) {
	record := &fileUrlRecord{
		Uuid:        u.Length(),
		ShortURL:    key,
		OriginalURL: url,
	}

	u.urls[key] = url

	if err := u.encoder.Encode(&record); err != nil {
		panic(err)
	}
}
func (u *FileStorage) Length() int {
	return len(u.urls) + 1
}

func (u *FileStorage) Close() error {
	return u.element.Close()
}
