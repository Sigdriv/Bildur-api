package db

import (
	"fmt"

	"github.com/Sigdriv/Bildur-api/model"
	"github.com/google/uuid"
)

func (db *DB) GetImages() (images []model.PreviewImage, err error) {
	query := `
	select ip.id, ip."originalImageId", ip."storagePath", ip.width, ip.height, ip."createdAt", i."extension", i."name"
	from "imagePreviews" ip
	left join images i on i.id = ip."originalImageId"
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
	select i.id, i."name", i."mimeType", i.bytes, i."storagePath", i.width, i.height, i."createdAt", i.extension, g.id as "greyScaleId"
	from images i 
	left join "greyScaleImages" g on g."originalImageId" = i.id
	where i.id = :id
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

func (db *DB) GetImagesByIDs(IDs []uuid.UUID) (images []model.Image, err error) {
	query := `
	select i.id, i."name", i."mimeType", i.bytes, i."storagePath", i.width, i.height, i."createdAt", i.extension
	from images i 
	where i.id IN (:ids)
	`

	args := map[string]any{
		"ids": IDs,
	}

	query, args = In(query, args)

	images, err = Query[model.Image](db, query, args)
	if err != nil {
		err = fmt.Errorf("error fetching images by IDs from database >> %s", err)
		return
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

func (db *DB) InsertThumbnailImage(image model.InsertThumbnailImage) (id string, err error) {
	query := `
	insert into "imagePreviews" ("originalImageId", "storagePath", width, height, "createdAt")
	values (:originalImageId, :storagePath, :width, :height, :createdAt)
	returning id
	`

	args := map[string]any{
		"originalImageId": image.OriginalImageId,
		"storagePath":     image.StoragePath,
		"width":           image.Width,
		"height":          image.Height,
		"createdAt":       image.CreatedAt,
	}

	query, args = In(query, args)

	id, err = Exec(db, query, args)
	if err != nil {
		err = fmt.Errorf("error inserting thumbnail image into database >> %s", err)
		return
	}

	return
}

func (db *DB) InsertGreyScaleImage(image model.InsertGreyScaleImage) (id uuid.UUID, err error) {
	query := `
	insert into "greyScaleImages" ("originalImageId", "storagePath", "createdAt")
	values (:originalImageId, :storagePath, :createdAt)
	returning id
	`

	args := map[string]any{
		"originalImageId": image.OriginalImageId,
		"storagePath":     image.StoragePath,
		"createdAt":       image.CreatedAt,
	}

	query, args = In(query, args)

	idStr, err := Exec(db, query, args)
	if err != nil {
		err = fmt.Errorf("error inserting grey scale image into database >> %s", err)
		return
	}

	id, err = uuid.Parse(idStr)
	if err != nil {
		err = fmt.Errorf("error parsing grey scale image ID >> %s", err)
		return
	}

	return
}
