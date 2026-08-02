package immich

type Asset struct {
	ID               string `json:"id"`
	OriginalFileName string `json:"originalFileName"`
	OriginalPath     string `json:"originalPath"`
}

type StackResponse struct {
	Assets         []Asset `json:"assets"`
	ID             string  `json:"id"`
	PrimaryAssetId string  `json:"primaryAssetId"`
}
