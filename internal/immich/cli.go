package immich

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
)

type CliAsset struct {
	Id       string `json:"id"`
	Filepath string `json:"filepath"`
}

type UploadResponse struct {
	NewFiles   []string   `json:"newFiles"`
	Duplicates []CliAsset `json:"duplicates"`
	NewAssets  []CliAsset `json:"newAssets"`
}

func LoginWithImmichCli(endpoint, apiKey string) (string, string, error) {
	cmd := exec.Command("immich", "login", endpoint, apiKey)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return stdout.String(), stderr.String(), fmt.Errorf("immich login failed: %w", err)
	}
	return stdout.String(), stderr.String(), nil
}

func UploadWithImmichCli(filePath string) (*UploadResponse, error) {
	cmd := exec.Command("immich", "upload", filePath+"/*", "--album", "--json-output", "--no-progress")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("immich upload failed: %w, output: %s", err, string(output))
	}

	uploadResponse, err := ExtractJSON(bytes.NewReader(output))
	if err != nil {
		return nil, fmt.Errorf("failed to parse upload output as JSON: %w, output: %s", err, string(output))
	}
	return uploadResponse, nil
}

func ExtractJSON(r io.Reader) (*UploadResponse, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}

	start := bytes.IndexByte(data, '{')
	if start == -1 {
		return nil, fmt.Errorf("json not found")
	}

	decoder := json.NewDecoder(bytes.NewReader(data[start:]))

	var v UploadResponse
	if err := decoder.Decode(&v); err != nil {
		return nil, err
	}
	return &v, nil
}
