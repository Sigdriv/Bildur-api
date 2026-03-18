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

func (db *DB) GetImageByID(id string) (image *model.Image, err error) {
	query := `
	select i.id, i."name", i."mimeType", i.bytes, i."storagePath", i.width, i.height, i."createdAt", i.extension
	from images i 
	where id = :id
	`

	args := map[string]any{
		"id": id,
	}

	images, err := Query[model.Image](db, query, args)
	if err != nil {
		err = fmt.Errorf("error fetching image from database >> %s", err)
		return
	}

	if len(images) > 0 {
		image = &images[0]
	}

	return
}

func (db *DB) InsertImage(image model.InsertImage) (id string, err error) {
	query := `
	insert into images (name, "mimeType", extension, bytes, "storagePath", width, height)
	values (:name, :mimeType, :extension, :bytes, :storagePath, :width, :height)
	returning id
	`

	args := map[string]any{
		"name":        image.Name,
		"mimeType":    image.MimeType,
		"extension":   image.Extension,
		"bytes":       image.Bytes,
		"storagePath": image.StoragePath,
		"width":       image.Width,
		"height":      image.Height,
	}

	query, args = In(query, args)

	id, err = Exec(db, query, args)
	if err != nil {
		err = fmt.Errorf("error inserting image into database >> %s", err)
		return
	}

	return
}
