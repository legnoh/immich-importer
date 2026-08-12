package immich

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type ImmichClient struct {
	Endpoint   string
	ApiKey     string
	HttpClient http.Client
}

func NewImmichClient(endpoint string, apiKey string) (*ImmichClient, error) {
	return &ImmichClient{
		Endpoint:   endpoint,
		ApiKey:     apiKey,
		HttpClient: http.Client{},
	}, nil
}

func (c *ImmichClient) request(method, path string, body []byte) ([]byte, error) {
	endpoint := strings.TrimRight(c.Endpoint, "/") + path
	req, err := http.NewRequest(method, endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to build request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", c.ApiKey)

	resp, err := c.HttpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call immich endpoint: %w", err)
	}
	if resp.StatusCode >= 300 {
		responseBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("immich create stack failed: status=%d body=%s", resp.StatusCode, string(responseBody))
	}
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}
	return responseBody, nil
}

func (c *ImmichClient) CreateStack(assetIds []string) (*StackResponse, error) {
	if len(assetIds) == 0 {
		return nil, fmt.Errorf("assetIds is required")
	}

	payload := struct {
		AssetIds []string `json:"assetIds"`
	}{
		AssetIds: assetIds,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	responseBody, err := c.request(http.MethodPost, "/api/stacks", body)
	if err != nil {
		return nil, fmt.Errorf("failed to call immich create stack endpoint: %w", err)
	}

	var stackResponse StackResponse
	if err := json.Unmarshal(responseBody, &stackResponse); err != nil {
		return nil, fmt.Errorf("failed to decode create stack response: %w", err)
	}
	return &stackResponse, nil
}

func (c *ImmichClient) UpdateStack(stackId string, primaryAssetId string) (bool, error) {
	if stackId == "" || primaryAssetId == "" {
		return false, fmt.Errorf("stackId and primaryAssetId are required")
	}

	payload := struct {
		PrimaryAssetId string `json:"primaryAssetId"`
	}{
		PrimaryAssetId: primaryAssetId,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return false, fmt.Errorf("failed to marshal request body: %w", err)
	}

	if _, err := c.request(http.MethodPut, "/api/stacks/"+stackId, body); err != nil {
		return false, fmt.Errorf("failed to call immich update stack endpoint: %w", err)
	}
	return true, nil
}

func (c *ImmichClient) GetAlbums() ([]*Album, error) {
	responseBody, err := c.request(http.MethodGet, "/api/albums", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to call immich get albums endpoint: %w", err)
	}

	var albumsResponse []*Album
	if err := json.Unmarshal(responseBody, &albumsResponse); err != nil {
		return nil, fmt.Errorf("failed to decode get albums response: %w", err)
	}
	return albumsResponse, nil
}

func (c *ImmichClient) CreateAlbum(name string, assetIds []string) (*Album, error) {
	if name == "" {
		return nil, fmt.Errorf("name is required")
	}

	payload := struct {
		Name     string   `json:"albumName"`
		AssetIds []string `json:"assetIds"`
	}{
		Name:     name,
		AssetIds: assetIds,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	responseBody, err := c.request(http.MethodPost, "/api/albums", body)
	if err != nil {
		return nil, fmt.Errorf("failed to call immich create album endpoint: %w", err)
	}

	var album Album
	if err := json.Unmarshal(responseBody, &album); err != nil {
		return nil, fmt.Errorf("failed to decode create album response: %w", err)
	}
	return &album, nil
}

func (c *ImmichClient) UpdateAlbum(albumId string, albumThumbnailAssetId string) (bool, error) {
	if albumId == "" || albumThumbnailAssetId == "" {
		return false, fmt.Errorf("albumId and albumThumbnailAssetId are required")
	}

	payload := struct {
		AlbumThumbnailAssetId string `json:"albumThumbnailAssetId"`
		Order                 string `json:"order"`
	}{
		AlbumThumbnailAssetId: albumThumbnailAssetId,
		Order:                 "asc", // アルバムの順序を降順に設定
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return false, fmt.Errorf("failed to marshal request body: %w", err)
	}

	if _, err := c.request(http.MethodPatch, "/api/albums/"+albumId, body); err != nil {
		return false, fmt.Errorf("failed to call immich update album endpoint: %w", err)
	}
	return true, nil
}

func (c *ImmichClient) GetAlbumAssets(albumIds []string) ([]Asset, error) {
	if len(albumIds) == 0 {
		return nil, fmt.Errorf("albumIds are required")
	}

	requestBody := struct {
		AlbumIds []string `json:"albumIds"`
	}{
		AlbumIds: albumIds,
	}

	body, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	responseBody, err := c.request(http.MethodPost, "/api/search/metadata", body)
	if err != nil {
		return nil, fmt.Errorf("failed to call immich get album assets endpoint: %w", err)
	}

	var searchResponse SearchResponse
	if err := json.Unmarshal(responseBody, &searchResponse); err != nil {
		return nil, fmt.Errorf("failed to decode get album assets response: %w", err)
	}
	return searchResponse.Assets.Items, nil
}

func (c *ImmichClient) GetAssetInfo(assetId string) (*AssetResponse, error) {
	if assetId == "" {
		return nil, fmt.Errorf("assetId is required")
	}
	responseBody, err := c.request(http.MethodGet, "/api/assets/"+assetId, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to call immich get asset info endpoint: %w", err)
	}

	var asset AssetResponse
	if err := json.Unmarshal(responseBody, &asset); err != nil {
		return nil, fmt.Errorf("failed to decode get asset info response: %w", err)
	}
	return &asset, nil
}

func (c *ImmichClient) UpdateAsset(assetId string, dateTimeOriginal time.Time) (bool, error) {
	if assetId == "" {
		return false, fmt.Errorf("assetId is required")
	}

	payload := struct {
		DateTimeOriginal string `json:"dateTimeOriginal,omitempty"`
		Description      string `json:"description,omitempty"`
		IsFavorite       bool   `json:"isFavorite,omitempty"`
		Latitude         uint   `json:"latitude,omitempty"`
		LivePhotoVideoId string `json:"livePhotoVideoId,omitempty"`
		Longitude        uint   `json:"longitude,omitempty"`
		Rating           int    `json:"rating,omitempty"`
		Visibility       string `json:"visibility,omitempty"`
	}{
		DateTimeOriginal: dateTimeOriginal.Format("2006-01-02T15:04:05.000Z"),
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return false, fmt.Errorf("failed to marshal request body: %w", err)
	}

	if _, err := c.request(http.MethodPut, "/api/assets/"+assetId, body); err != nil {
		return false, fmt.Errorf("failed to call immich update asset info endpoint: %w", err)
	}
	return true, nil
}
