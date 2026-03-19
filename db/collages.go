package db

import (
	"fmt"

	"github.com/google/uuid"
)

func (db *DB) InsertImagesToCollages(imageIDs []uuid.UUID, storagePath string, cols, rows int) (collageID string, err error) {

	query := `
	insert into collages ("rows", "cols", "storagePath")
	values (:rows, :cols, :storagePath)
	returning id
	`

	args := map[string]any{
		"rows":        rows,
		"cols":        cols,
		"storagePath": storagePath,
	}

	collageID, err = Exec(db, query, args)
	if err != nil {
		err = fmt.Errorf("error inserting collage into database >> %s", err)
		return
	}

	insertImagesQuery := `
	insert into "collageImages" ("collageId", "imageId", "rowIndex", "colIndex")
	values (:collageId, :imageId, :rowIndex, :colIndex)
	`

	for i, imageID := range imageIDs {
		rowIndex := i / cols
		colIndex := i % cols

		args := map[string]any{
			"collageId": collageID,
			"imageId":   imageID,
			"rowIndex":  rowIndex,
			"colIndex":  colIndex,
		}

		_, err = Exec(db, insertImagesQuery, args)
		if err != nil {
			err = fmt.Errorf("error inserting image into collage >> %s", err)
			return
		}
	}

	return
}

func (db *DB) UpdateCollage(collageID string, width, height int) (err error) {
	query := `
	update collages
	set width = :width, height = :height
	where id = :id
	`

	args := map[string]any{
		"id":     collageID,
		"width":  width,
		"height": height,
	}

	_, err = Exec(db, query, args)
	if err != nil {
		err = fmt.Errorf("error updating collage dimensions >> %s", err)
		return
	}

	return
}
