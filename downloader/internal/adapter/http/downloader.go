package http

import (
	"context"
	"io"
	"net/http"
	"time"
)

type HTTPFileDownloader struct {
	client *http.Client
}

func NewHTTPFileDownloader(timeout time.Duration) *HTTPFileDownloader {
	return &HTTPFileDownloader{
		client: &http.Client{Timeout: timeout},
	}
}

func (d *HTTPFileDownloader) Download(ctx context.Context, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := d.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return data, nil
}
