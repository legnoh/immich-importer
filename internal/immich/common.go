package immich

type Asset struct {
	ID               string `json:"id"`
	OriginalFileName string `json:"originalFileName"`
	OriginalPath     string `json:"originalPath"`
	Width            int    `json:"width"`
	Height           int    `json:"height"`
}

type Album struct {
	ID                         string `json:"id"`
	Name                       string `json:"albumName"`
	Description                string `json:"description"`
	AlbumThumbnailAssetId      string `json:"albumThumbnailAssetId"`
	CreatedAt                  string `json:"createdAt"`
	UpdatedAt                  string `json:"updatedAt"`
	AssetCount                 int    `json:"assetCount"`
	IsActivityEnabled          bool   `json:"isActivityEnabled"`
	LastModifiedAssetTimestamp string `json:"lastModifiedAssetTimestamp"`
	Order                      string `json:"order"`
}

type StackResponse struct {
	ID             string  `json:"id"`
	PrimaryAssetId string  `json:"primaryAssetId"`
	Assets         []Asset `json:"assets"`
}

type AssetResponse struct {
	ID               string `json:"id"`
	OriginalFileName string `json:"originalFileName"`
	OriginalPath     string `json:"originalPath"`
	Width            int    `json:"width"`
	Height           int    `json:"height"`
	FileCreatedAt    string `json:"fileCreatedAt"`
	FileModifiedAt   string `json:"fileModifiedAt"`
	LocalDateTime    string `json:"localDateTime"`
	Stack            struct {
		ID             string `json:"id"`
		PrimaryAssetId string `json:"primaryAssetId"`
		AssetCount     int    `json:"assetCount"`
	} `json:"stack"`
}

type SearchResponse struct {
	Albums struct {
		Total int     `json:"total"`
		Count int     `json:"count"`
		Items []Album `json:"items"`
	} `json:"albums"`
	Assets struct {
		Total int     `json:"total"`
		Count int     `json:"count"`
		Items []Asset `json:"items"`
	} `json:"assets"`
}
