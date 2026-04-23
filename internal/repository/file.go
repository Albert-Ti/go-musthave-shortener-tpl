package repository

import (
	"encoding/json"
	"os"
)

type FileRepository struct {
	File    *os.File
	Encoder *json.Encoder
}

type FileUrlRecord struct {
	Uuid        uint   `json:"uuid"`
	ShortUrl    string `json:"short_url"`
	OriginalUrl string `json:"original_url"`
}

func NewFileRepository(filename string) (*FileRepository, error) {
	file, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}

	return &FileRepository{
		File:    file,
		Encoder: json.NewEncoder(file),
	}, nil
}

func (f *FileRepository) WriteUrl(fileUrlRecord *FileUrlRecord) error {
	return f.Encoder.Encode(fileUrlRecord)
}
