package immich

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"

	"gopkg.in/yaml.v3"
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

func GetImmichCliLoginInfo() (string, string, error) {

	// $HOME/.config/immich/auth.ymlがあるか確認し、あればそれを自動で読み込む
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", "", fmt.Errorf("failed to get user home directory: %w", err)
	}
	authFilePath := filepath.Join(homeDir, ".config", "immich", "auth.yml")
	if _, err := os.Stat(authFilePath); os.IsNotExist(err) {
		return "", "", fmt.Errorf("immich auth file not found: %s", authFilePath)
	}

	// auth.ymlをYAMLとして読み込む
	file, err := os.Open(authFilePath)
	if err != nil {
		return "", "", fmt.Errorf("failed to open immich auth file: %w", err)
	}
	defer file.Close()

	var authData struct {
		Url string `yaml:"url"`
		Key string `yaml:"key"`
	}
	decoder := yaml.NewDecoder(file)
	if err := decoder.Decode(&authData); err != nil {
		return "", "", fmt.Errorf("failed to parse immich auth file: %w", err)
	}

	// urlは"/api"がついているはずなので、"api"の部分を削除する
	endpoint := authData.Url
	if len(endpoint) >= 4 && endpoint[len(endpoint)-4:] == "/api" {
		endpoint = endpoint[:len(endpoint)-3]
	}
	return endpoint, authData.Key, nil
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

func UploadWithImmichCli(files []string) (*UploadResponse, error) {
	cmd := exec.Command("immich", "upload")
	cmd.Args = append(cmd.Args, files...)
	cmd.Args = append(cmd.Args, "--json-output", "--no-progress")
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
