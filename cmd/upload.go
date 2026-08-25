package cmd

import (
	"path/filepath"
	"sort"
	"time"

	"github.com/legnoh/immich-importer/internal/immich"
	"github.com/legnoh/immich-importer/internal/logger"
)

type UploadCmd struct {
	FolderPath string `arg:"" name:"path" help:"Path to the file to upload."`
}

func (c *UploadCmd) Run(g GlobalFlags) error {
	log := logger.Default

	// すでにimmich cliでログイン済みか確認し、自動でAPIキーとURLを読み込む
	if g.ImmichCliAutoLogin {
		log.Info("checking immich cli login status...")
		endpoint, apiKey, err := immich.GetImmichCliLoginInfo()
		if err != nil {
			log.Error("immich cli login info not found", "msg", err)
			return err
		}
		g.ImmichEndpoint = endpoint
		g.ImmichApiKey = apiKey
		log.Info("immich cli login info found", "endpoint", g.ImmichEndpoint)
	}

	// パス配下のファイルを取得する
	files, err := filepath.Glob(c.FolderPath + "/*")
	if err != nil {
		log.Error("failed to get files", "msg", err)
		return err
	}

	//ファイル名でソートする
	sort.Slice(files, func(i, j int) bool {
		return filepath.Base(files[i]) < filepath.Base(files[j])
	})

	// immich cliでファイルをアップロードする
	log.Info("uploading files with immich cli...", "fileCount", len(files))
	uploadResponse, err := immich.UploadWithImmichCli(files)
	if err != nil {
		log.Error("failed to upload with immich cli", "msg", err)
		return err
	}
	log.Info("upload completed", "duplicates", len(uploadResponse.Duplicates), "newAssets", len(uploadResponse.NewAssets))

	// http client作成
	client, err := immich.NewImmichClient(g.ImmichEndpoint, g.ImmichApiKey)
	if err != nil {
		log.Error("failed to create immich client", "msg", err)
		return err
	}

	// 一旦2つの配列を結合して、Album/Stack作成用のassetIdsを作り、ファイル名でソートする
	assets := append(uploadResponse.NewAssets, uploadResponse.Duplicates...)
	sort.Slice(assets, func(i, j int) bool {
		return filepath.Base(assets[i].Filepath) < filepath.Base(assets[j].Filepath)
	})

	// Album/Stack作成用のassetIdsを作る
	var assetIds []string
	for _, asset := range assets {
		assetIds = append(assetIds, asset.Id)
	}

	// 先頭から縦長のアセットを探して、アルバムのサムネイル用にIDを保管する
	var primaryAsset *immich.AssetResponse = nil
	for _, asset := range assets {
		asset, err := client.GetAssetInfo(asset.Id)
		if err != nil {
			log.Error("failed to get asset info", "msg", err)
			return err
		}
		if asset.Height > asset.Width || asset.Height == asset.Width {
			primaryAsset = asset
			break
		}
	}

	// Album/Stackを作る
	if len(uploadResponse.NewAssets) > 0 || len(uploadResponse.Duplicates) > 0 {

		// Albumを作成する
		album, err := client.CreateAlbum(filepath.Base(c.FolderPath), assetIds)
		if err != nil {
			log.Error("failed to create album", "msg", err)
			return err
		}
		log.Info("album created successfully", "albumName", album.Name, "url", g.ImmichEndpoint+"/albums/"+album.ID)

		// Albumのサムネイルを1番目のアセットにして、昇順に並べる
		_, err = client.UpdateAlbum(album.ID, primaryAsset.ID)
		if err != nil {
			log.Error("failed to set album thumbnail", "msg", err)
			return err
		}

		// Stackを作成する
		stackResponse, err := client.CreateStack(assetIds)
		if err != nil {
			log.Error("failed to create stack", "msg", err)
			return err
		}
		log.Info("stack created successfully", "stackId", stackResponse.ID, "url", g.ImmichEndpoint+"/stacks/"+stackResponse.ID)

		// Stackのメインを1番目のアセットにする
		_, err = client.UpdateStack(stackResponse.ID, primaryAsset.ID)
		if err != nil {
			log.Error("failed to set stack main asset", "msg", err)
			return err
		}

		// プライマリAssetの撮影日時を先頭に、全Assetの撮影日時を昇順に更新していく
		// まずは基準となる撮影日時を取得し、0時0分に設定する
		dateTimeStr := primaryAsset.LocalDateTime
		dateTimeOrigin, err := time.Parse("2006-01-02T15:04:05.000Z", dateTimeStr)
		if err != nil {
			log.Error("failed to parse primary asset date time", "msg", err, "assetId", primaryAsset.ID)
			return err
		}
		dateTimeBase := time.Date(dateTimeOrigin.Year(), dateTimeOrigin.Month(), dateTimeOrigin.Day(), 0, 0, 0, 0, dateTimeOrigin.Location())

		// 先ほどのソートしたアルバムアセットを順番に更新していく
		log.Info("updating asset date time in ascending order...", "assetCount", len(assets))
		for i, asset := range assets {
			newDateTime := dateTimeBase.Add(time.Duration(i) * time.Second)
			log.Debug("updating asset date time", "assetId", asset.Id, "newDateTime", newDateTime.Format("2006-01-02T15:04:05.000Z"))
			_, err = client.UpdateAsset(asset.Id, newDateTime)
			if err != nil {
				log.Error("failed to update asset date time", "msg", err, "assetId", asset.Id)
				return err
			}
		}
	}
	log.Info("all upload process was completed successfully!")
	return nil
}
