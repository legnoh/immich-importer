package cmd

import (
	"path/filepath"
	"sort"
	"time"

	"github.com/legnoh/immich-importer/internal/immich"
	"github.com/legnoh/immich-importer/internal/logger"
)

type UpdateCmd struct {
}

func (c *UpdateCmd) Run(g GlobalFlags) error {
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

	// client作成
	client, err := immich.NewImmichClient(g.ImmichEndpoint, g.ImmichApiKey)
	if err != nil {
		log.Error("failed to create immich client", "msg", err)
		return err
	}

	// アルバム一覧を取得する
	albums, err := client.GetAlbums()
	if err != nil {
		log.Error("failed to get albums", "msg", err)
		return err
	}

	// 全てのアルバムに対して以下を実行する
	for _, album := range albums {

		log.Info("updating album", "albumId", album.ID, "albumName", album.Name)

		// アルバムのアセットを取得する
		albumAssets, err := client.GetAlbumAssets([]string{album.ID})
		if err != nil {
			log.Error("failed to get album assets", "msg", err, "albumId", album.ID)
			return err
		}

		// アセットをファイル名でソートする
		sort.Slice(albumAssets, func(i, j int) bool {
			return filepath.Base(albumAssets[i].OriginalFileName) < filepath.Base(albumAssets[j].OriginalFileName)
		})

		// 先頭から縦長のアセットを探して、アルバムのサムネイル用にIDを保管する
		var primaryAssetId string = ""
		for _, asset := range albumAssets {
			if asset.Height > asset.Width || asset.Height == asset.Width {
				primaryAssetId = asset.ID
				break
			}
		}

		// 先頭のAssetが取得できなかった場合は処理をスキップする
		if primaryAssetId == "" {
			log.Warn("no suitable primary asset found for album", "albumId", album.ID)
			continue
		}

		// アルバムのサムネイルに設定する
		if len(albumAssets) > 0 {
			_, err = client.UpdateAlbum(album.ID, primaryAssetId)
			if err != nil {
				log.Error("failed to update album thumbnail", "msg", err, "albumId", album.ID)
				return err
			}
		}

		// 先頭のAsset情報を再度取得
		primAsset, err := client.GetAssetInfo(albumAssets[0].ID)
		if err != nil {
			log.Error("failed to get primary asset info", "msg", err, "assetId", albumAssets[0].ID)
			return err
		}

		// Stackに所属しているか確認
		if primAsset.Stack.ID == "" {

			log.Info("primary asset is not in a stack, creating stack", "assetId", primaryAssetId)

			// ない場合はStackを作る
			var assetIds []string
			for _, asset := range albumAssets {
				assetIds = append(assetIds, asset.ID)
			}
			stackResponse, err := client.CreateStack(assetIds)
			if err != nil {
				log.Error("failed to create stack", "msg", err, "albumId", album.ID)
				return err
			}
			primAsset.Stack.ID = stackResponse.ID
		}

		// 先ほど取得したアルバムカバーをアセットのメイン画像にする
		if len(albumAssets) > 0 {
			_, err = client.UpdateStack(primAsset.Stack.ID, primaryAssetId)
			if err != nil {
				log.Error("failed to set stack main asset", "msg", err, "stackId", primaryAssetId)
				return err
			}
		}

		// プライマリAssetの撮影日時を先頭に、全Assetの撮影日時を昇順に更新していく
		// まずは基準となる撮影日時を取得し、0時0分に設定する
		dateTimeStr := primAsset.LocalDateTime
		dateTimeOrigin, err := time.Parse("2006-01-02T15:04:05.000Z", dateTimeStr)
		if err != nil {
			log.Error("failed to parse primary asset date time", "msg", err, "assetId", primaryAssetId)
			return err
		}
		dateTimeBase := time.Date(dateTimeOrigin.Year(), dateTimeOrigin.Month(), dateTimeOrigin.Day(), 0, 0, 0, 0, dateTimeOrigin.Location())

		// 先ほどのソートしたアルバムアセットを順番に更新していく
		for i, asset := range albumAssets {
			newDateTime := dateTimeBase.Add(time.Duration(i) * time.Second)
			log.Debug("updating asset date time", "assetId", asset.ID, "newDateTime", newDateTime.Format("2006-01-02T15:04:05.000Z"))
			_, err = client.UpdateAsset(asset.ID, newDateTime)
			if err != nil {
				log.Error("failed to update asset date time", "msg", err, "assetId", asset.ID)
				return err
			}
		}

		log.Info("album updated successfully", "albumId", album.ID, "albumName", album.Name)
	}
	log.Info("albums updated successfully")
	return nil
}
