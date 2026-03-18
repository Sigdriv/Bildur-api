package db

import (
	"fmt"

	"github.com/Sigdriv/Bildur-api/model"
)

func (db *DB) GetImages() (images []model.PreviewImage, err error) {
	query := `
	select ip.id, ip."imageId", ip."variantName", ip."storagePath", ip.width, ip.height, ip."createdAt"
	from "imagePreviews" ip
	`

	images, err = Query[model.PreviewImage](db, query, nil)
	if err != nil {
		err = fmt.Errorf("error fetching images from database >> %s", err)
		return
	}

	return
}
