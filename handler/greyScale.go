package handler

import (
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/Sigdriv/Bildur-api/db"
	"github.com/Sigdriv/Bildur-api/model"
	"github.com/gin-gonic/gin"
)

func (srv *Handler) HandleGreyScaleImage(c *gin.Context) {
	log := srv.getLog(c)

	id := c.Param("id")
	if id == "" {
		log.Warn("No image ID provided for grey scale conversion")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Image ID is required"})
		return
	}

	image, err := srv.DB.GetImageByID(id)
	if err != nil {
		log.Errorf("Error fetching image from database >> %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch image"})
		return
	}

	if image == nil {
		log.Warnf("Requested image not found for grey scale conversion: %s", id)
		c.JSON(http.StatusNotFound, gin.H{"error": "Image not found"})
		return
	}

	if image.GreyScaleID != nil {
		log.Infof("Image already has a grey scale version, skipping conversion for image ID: %s", id)
		c.JSON(http.StatusOK, gin.H{"message": "Image already has a grey scale version"})
		return
	}

	err = convertToGreyScale(image, srv.DB)
	if err != nil {
		log.Errorf("Error converting image to grey scale >> %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to convert image to grey scale"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Image converted to grey scale successfully"})
}

func convertToGreyScale(img *model.Image, db db.DB) (err error) {
	filePath := fmt.Sprintf("./%s", FullSizeStoragePath)

	storageFileName := fmt.Sprintf("%s.%s", img.ID, img.Extension)
	storagePath := filepath.Join(filePath, storageFileName)

	inFile, err := os.Open(storagePath)
	if err != nil {
		err = fmt.Errorf("error opening image file >> %v", err)
		return
	}
	defer inFile.Close()

	decodedImg, _, err := image.Decode(inFile)
	if err != nil {
		err = fmt.Errorf("error decoding image file >> %v", err)
		return
	}

	bounds := decodedImg.Bounds()
	greyImg := image.NewGray(bounds)

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			rgba := color.RGBAModel.Convert(decodedImg.At(x, y)).(color.RGBA)

			greyVal := uint8(0.299*float64(rgba.R) + 0.587*float64(rgba.G) + 0.114*float64(rgba.B))
			greyImg.Set(x, y, color.Gray{Y: greyVal})
		}
	}

	greyScaleImage := model.InsertGreyScaleImage{
		OriginalImageId: img.ID,
		StoragePath:     GreyScaleStoragePath,
		CreatedAt:       time.Now(),
	}

	greyScaleID, err := db.InsertGreyScaleImage(greyScaleImage)
	if err != nil {
		err = fmt.Errorf("error inserting grey scale image into database >> %v", err)
		return
	}

	dstPath := fmt.Sprintf("./%s", GreyScaleStoragePath)
	err = createPathIfNotExists(dstPath)
	if err != nil {
		err = fmt.Errorf("error creating thumbnail directory >> %w", err)
		return
	}

	greyScaleFileName := fmt.Sprintf("%s.%s", greyScaleID, img.Extension)
	greyScalePath := filepath.Join(dstPath, greyScaleFileName)

	outFile, err := os.Create(greyScalePath)
	if err != nil {
		err = fmt.Errorf("error creating grey scale image file >> %v", err)
		return
	}
	defer outFile.Close()

	switch img.Extension {
	case "png":
		err = png.Encode(outFile, greyImg)
	case "jpg", "jpeg":
		err = jpeg.Encode(outFile, greyImg, nil)
	default:
		err = fmt.Errorf("unsupported image format for grey scale conversion: %s", img.Extension)
	}

	return
}
