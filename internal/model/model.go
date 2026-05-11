package model

type JSONReq struct {
	URL string `json:"url"`
}

type JSONResp struct {
	Result string `json:"result"`
}

type JSONBatchReq struct {
	CorrelationID string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
}

type JSONBatchResp struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}
