package immich

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type ImmichClient struct {
	Endpoint   string
	ApiKey     string
	HttpClient http.Client
}

func NewImmichClient(endpoint, apiKey string) (*ImmichClient, error) {
	return &ImmichClient{
		Endpoint:   endpoint,
		ApiKey:     apiKey,
		HttpClient: http.Client{},
	}, nil
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

	endpoint := strings.TrimRight(c.Endpoint, "/") + "/api/stacks"
	req, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to build request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", c.ApiKey)

	resp, err := c.HttpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call immich create stack endpoint: %w", err)
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("immich create stack failed: status=%d body=%s", resp.StatusCode, string(responseBody))
	}

	var stackResponse StackResponse
	if err := json.Unmarshal(responseBody, &stackResponse); err != nil {
		return nil, fmt.Errorf("failed to decode create stack response: %w", err)
	}

	return &stackResponse, nil
}

func (c *ImmichClient) UpdateStack(stackId string, primaryAssetId string) (bool, error) {
	if stackId == "" {
		return false, fmt.Errorf("stackId is required")
	}
	if primaryAssetId == "" {
		return false, fmt.Errorf("primaryAssetId is required")
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

	endpoint := strings.TrimRight(c.Endpoint, "/") + "/api/stacks/" + stackId
	req, err := http.NewRequest(http.MethodPut, endpoint, bytes.NewReader(body))
	if err != nil {
		return false, fmt.Errorf("failed to build request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", c.ApiKey)

	resp, err := c.HttpClient.Do(req)
	if err != nil {
		return false, fmt.Errorf("failed to call immich update stack endpoint: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		responseBody, _ := io.ReadAll(resp.Body)
		return false, fmt.Errorf("immich update stack failed: status=%d body=%s", resp.StatusCode, string(responseBody))
	}
	return true, nil
}
