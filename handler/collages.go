package handler

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"math"
	"net/http"
	"os"
	"path/filepath"

	"github.com/Sigdriv/Bildur-api/model"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/image/draw"
)

func (srv *Handler) HandleCreateCollage(c *gin.Context) {
	log := srv.getLog(c)

	var rawIDs []string
	err := c.ShouldBindJSON(&rawIDs)
	if err != nil {
		log.Errorf("Error binding JSON for collage creation >> %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if len(rawIDs) == 0 {
		log.Warn("No image IDs provided for collage creation")
		c.JSON(http.StatusBadRequest, gin.H{"error": "At least one image ID is required"})
		return
	}

	imageIDs := make([]uuid.UUID, 0, len(rawIDs))
	for _, s := range rawIDs {
		id, err := uuid.Parse(s)
		if err != nil {
			log.Errorf("Error parsing image ID '%s' for collage creation >> %v", s, err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid image ID format", "value": s})
			return
		}

		imageIDs = append(imageIDs, id)
	}

	images, err := srv.DB.GetImagesByIDs(imageIDs)
	if err != nil {
		log.Errorf("Error fetching images for collage creation >> %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch images for collage"})
		return
	}

	if len(images) != len(imageIDs) {
		log.Warnf("Some image IDs not found for collage creation. Requested: %d, Found: %d", len(imageIDs), len(images))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Some image IDs were not found", "requested": len(imageIDs), "found": len(images)})
		return
	}

	imageLength := len(imageIDs)
	if imageLength == 0 {
		log.Warn("No valid image IDs provided for collage creation after parsing")
		c.JSON(http.StatusBadRequest, gin.H{"error": "At least one valid image ID is required"})
		return
	}

	cols := int(math.Ceil(math.Sqrt(float64(imageLength))))
	rows := int(math.Ceil(float64(imageLength) / float64(cols)))

	collageID, err := srv.DB.InsertImagesToCollages(imageIDs, CollageStoragePath, cols, rows)
	if err != nil {
		log.Errorf("Error creating collage in database >> %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create collage"})
		return
	}

	collageBytes, err := buildCollageImage(images, cols, rows)
	if err != nil {
		log.Errorf("Error building collage image >> %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to build collage image"})
		return
	}

	err = saveCollageImage(collageID, collageBytes)
	if err != nil {
		log.Errorf("Error saving collage image >> %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save collage image"})
		return
	}

	width, height, err := getImageDimensionsFromBytes(collageBytes)
	if err != nil {
		log.Errorf("Error decoding collage dimensions >> %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read collage dimensions"})
		return
	}

	err = srv.DB.UpdateCollage(collageID, width, height)
	if err != nil {
		log.Errorf("Error updating collage dimensions in database >> %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update collage dimensions"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"id": collageID})
}

func getImageDimensionsFromBytes(imageBytes []byte) (width, height int, err error) {
	if len(imageBytes) == 0 {
		err = fmt.Errorf("image bytes are empty")
		return
	}

	cfg, _, err := image.DecodeConfig(bytes.NewReader(imageBytes))
	if err != nil {
		err = fmt.Errorf("error decoding image config from bytes >> %w", err)
		return
	}

	width = cfg.Width
	height = cfg.Height

	return
}

func buildCollageImage(images []model.Image, cols, rows int) (imageBytes []byte, err error) {
	loaded := make([]image.Image, 0, len(images))
	minW, minH := 0, 0

	for _, img := range images {
		imagePath := fmt.Sprintf("./%s/%s.%s", FullSizeStoragePath, img.ID, img.Extension)
		file, openErr := os.Open(imagePath)
		if openErr != nil {
			err = fmt.Errorf("error opening file at path '%s' >> %w", imagePath, openErr)
			return
		}
		defer file.Close()

		decoded, _, decodeErr := image.Decode(file)
		if decodeErr != nil {
			err = fmt.Errorf("error decoding image at path '%s' >> %w", imagePath, decodeErr)
			return
		}

		loaded = append(loaded, decoded)

		bounds := decoded.Bounds()
		w := bounds.Dx()
		h := bounds.Dy()

		if minW == 0 || w < minW {
			minW = w
		}
		if minH == 0 || h < minH {
			minH = h
		}
	}

	if minH == 0 || minW == 0 {
		err = fmt.Errorf("invalid minimum dimensions calculated for collage: minW=%d, minH=%d", minW, minH)
		return
	}

	canvasW := cols * minW
	canvasH := rows * minH

	canvas := image.NewRGBA(image.Rect(0, 0, canvasW, canvasH))

	draw.Draw(canvas, canvas.Bounds(), &image.Uniform{C: color.Black}, image.Point{}, draw.Src)

	for i, srcImg := range loaded {
		if 1 >= rows*cols {
			break
		}

		dst := image.NewRGBA(image.Rect(0, 0, minW, minH))
		draw.ApproxBiLinear.Scale(dst, dst.Bounds(), srcImg, srcImg.Bounds(), draw.Over, nil)

		row := i / cols
		col := i % cols

		x := col * minW
		y := row * minH
		cellRect := image.Rect(x, y, x+minW, y+minH)

		draw.Draw(canvas, cellRect, dst, image.Point{}, draw.Over)
	}

	var buf bytes.Buffer
	err = jpeg.Encode(&buf, canvas, nil)
	if err != nil {
		err = fmt.Errorf("error encoding collage image to JPEG >> %w", err)
		return
	}

	imageBytes = buf.Bytes()
	return
}

func saveCollageImage(collageID string, imageBytes []byte) (err error) {
	dstPath := fmt.Sprintf("./%s", CollageStoragePath)
	err = createPathIfNotExists(dstPath)
	if err != nil {
		err = fmt.Errorf("error creating collage directory >> %w", err)
		return
	}

	collageFileName := fmt.Sprintf("%s.jpeg", collageID)
	collagePath := filepath.Join(dstPath, collageFileName)

	outFile, err := os.Create(collagePath)
	if err != nil {
		err = fmt.Errorf("error creating collage image file >> %w", err)
		return
	}
	defer outFile.Close()

	_, err = outFile.Write(imageBytes)
	if err != nil {
		err = fmt.Errorf("error writing collage image to file >> %w", err)
		return
	}

	return
}
