package cmd

import (
	"path/filepath"
	"sort"

	"github.com/legnoh/immich-importer/internal/immich"
	"github.com/legnoh/immich-importer/internal/logger"
)

type UploadCmd struct {
	Path           string `arg:"" name:"path" help:"Path to the file to upload."`
	ImmichEndpoint string `help:"Immich endpoint URL." env:"IMC_ENDPOINT" default:"http://localhost:2283/"`
	ImmichApiKey   string `help:"Immich API key." env:"IMC_API_KEY" default:""`
}

func (c *UploadCmd) Run(g GlobalFlags) error {
	log := logger.Default

	// immich cliでログインする
	loginStdout, loginStderr, err := immich.LoginWithImmichCli(c.ImmichEndpoint, c.ImmichApiKey)
	if err != nil {
		log.Error("failed to login with immich cli", "msg", err, "stdout", loginStdout, "stderr", loginStderr)
		return err
	}
	log.Info("logged in with immich cli successfully", "endpoint", c.ImmichEndpoint, "stdout", loginStdout, "stderr", loginStderr)

	// immich cliでファイルをアップロードする
	uploadResponse, err := immich.UploadWithImmichCli(c.Path)
	if err != nil {
		log.Error("failed to upload with immich cli", "msg", err)
		return err
	}

	// 併せてStackもつくる
	if len(uploadResponse.NewAssets) > 0 || len(uploadResponse.Duplicates) > 0 {

		// 一旦2つの配列を結合して、Stack作成用のassetIdsを作り、ファイル名でソートする
		assets := append(uploadResponse.NewAssets, uploadResponse.Duplicates...)
		sort.Slice(assets, func(i, j int) bool {
			return filepath.Base(assets[i].Filepath) < filepath.Base(assets[j].Filepath)
		})

		// Stack作成用のassetIdsを作る
		var assetIds []string
		for _, asset := range assets {
			assetIds = append(assetIds, asset.Id)
		}

		client, err := immich.NewImmichClient(c.ImmichEndpoint, c.ImmichApiKey)
		if err != nil {
			log.Error("failed to create immich client", "msg", err)
			return err
		}

		// Stackを作成する
		stackResponse, err := client.CreateStack(assetIds)
		if err != nil {
			log.Error("failed to create stack", "msg", err)
			return err
		}

		// Stackのメインを1番目のアセットにする
		_, err = client.UpdateStack(stackResponse.ID, assetIds[0])
		if err != nil {
			log.Error("failed to set stack main asset", "msg", err)
			return err
		}
	}
	log.Info("file uploaded successfully", "duplicates", uploadResponse.Duplicates, "newAssets", uploadResponse.NewAssets)
	return nil
}
