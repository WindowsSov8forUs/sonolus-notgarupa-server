package repository

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type Client struct {
	BaseURL string
	client  *http.Client
}

type LevelUpload struct {
	Name        string   `json:"name"`
	Title       string   `json:"title"`
	Artists     string   `json:"artists"`
	Author      string   `json:"author"`
	Description string   `json:"description"`
	Tags        []string `json:"tags"`
	Rating      int      `json:"rating"`
	Engine      string   `json:"engine"`
	Cover       []byte   `json:"cover"`
	BGM         []byte   `json:"bgm"`
	Data        []byte   `json:"data"`
	Chart       []byte   `json:"chart,omitempty"`
}

type LevelUploadResponse struct {
	Version int64 `json:"version"`
}

func NewClient(baseURL string) *Client {
	return &Client{
		BaseURL: strings.TrimRight(baseURL, "/"),
		client:  &http.Client{Timeout: 60 * time.Second},
	}
}

func (c *Client) UploadLevel(ctx context.Context, upload LevelUpload) (LevelUploadResponse, error) {
	var response LevelUploadResponse
	data, err := json.Marshal(upload)
	if err != nil {
		return response, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.BaseURL+"/admin/levels", bytes.NewReader(data))
	if err != nil {
		return response, err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.client.Do(req)
	if err != nil {
		return response, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusConflict {
		return response, ErrNameCollision
	}
	if resp.StatusCode != http.StatusOK {
		return response, fmt.Errorf("repository upload failed: status=%d", resp.StatusCode)
	}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return response, err
	}
	return response, nil
}

var ErrNameCollision = fmt.Errorf("repository level name collision")
