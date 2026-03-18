package handler

import (
	"fmt"
	"image"
	_ "image/gif"
	"image/jpeg"
	_ "image/jpeg"
	"image/png"
	_ "image/png"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/Sigdriv/Bildur-api/db"
	"github.com/Sigdriv/Bildur-api/model"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/image/draw"
)

const FullSizeStoragePath = "media/fullsize"
const ThumbnailStoragePath = "media/thumbnails"
const MaxThumbnailSize = 300

func (srv *Handler) HandleGetImages(c *gin.Context) {
	log := srv.getLog(c)

	images, err := srv.DB.GetImages()
	if err != nil {
		log.Errorf("Error fetching images from database >> %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch images"})
	}

	if len(images) == 0 {
		images = []model.PreviewImage{}
	}

	c.JSON(http.StatusOK, gin.H{"images": images})
}

func (srv *Handler) HandleGetSingleImage(c *gin.Context) {
	log := srv.getLog(c)

	id := c.Param("id")
	image, err := srv.DB.GetImageByID(id)
	if err != nil {
		log.Errorf("Error fetching image from database >> %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch image"})
		return
	}

	if image == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Image not found"})
		return
	}

	c.JSON(http.StatusOK, image)
}

func (srv *Handler) HandleDownloadImage(c *gin.Context) {
	log := srv.getLog(c)

	id := c.Param("id")
	image, err := srv.DB.GetImageByID(id)
	if err != nil {
		log.Errorf("Error fetching image from database >> %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch image"})
		return
	}

	if image == nil {
		log.Warnf("Requested image not found: %s", id)
		c.JSON(http.StatusNotFound, gin.H{"error": "Image not found"})
		return
	}

	filePath := fmt.Sprintf("./%s/%s.%s", image.StoragePath, image.ID, image.Extension)

	c.Header("Content-Disposition", "attachment; filename="+image.Name)

	c.FileAttachment(filePath, image.Name)
}

func (srv *Handler) HandleUploadImage(c *gin.Context) {
	log := srv.getLog(c)

	file, fileHeader, err := c.Request.FormFile("file")
	if err != nil {
		log.Errorf("Error retrieving file from request >> %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to retrieve file"})
		return
	}
	defer file.Close()

	mimeType, ext, size, width, height, err := getFileInfo(file, fileHeader)
	if err != nil {
		log.Errorf("Error processing uploaded file >> %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to process file"})
		return
	}

	name := strings.Split(fileHeader.Filename, ".")[0]

	image := model.InsertImage{
		Name:        name,
		MimeType:    mimeType,
		Extension:   ext,
		Bytes:       size,
		StoragePath: FullSizeStoragePath,
		Width:       width,
		Height:      height,
	}

	id, err := srv.DB.InsertImage(image)
	if err != nil {
		log.Errorf("Error inserting image into database >> %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save image"})
		return
	}

	log.Infof("Successfully uploaded image with ID: %s", id)

	uploadDir := fmt.Sprintf("./%s", FullSizeStoragePath)
	err = createPathIfNotExists(uploadDir)
	if err != nil {
		log.Errorf("Error creating upload directory >> %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create upload directory"})
		return
	}

	storageFileName := fmt.Sprintf("%s.%s", id, ext)
	storagePath := filepath.Join(uploadDir, storageFileName)

	out, err := os.Create(storagePath)
	if err != nil {
		log.Errorf("Error creating file on disk >> %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file"})
		return
	}
	defer out.Close()

	_, err = io.Copy(out, file)
	if err != nil {
		log.Errorf("Error saving file to disk >> %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file"})
		return
	}

	err = createThumbnail(uuid.MustParse(id), ext, srv.DB)
	if err != nil {
		log.Errorf("Error creating thumbnail >> %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create thumbnail"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"id": id})
}

func getFileInfo(file multipart.File, fileHeader *multipart.FileHeader) (mimeType string, ext string, size int64, width int, height int, err error) {
	sniffSize := 512
	sniffBuffer := make([]byte, sniffSize)
	n, err := file.Read(sniffBuffer)
	if err != nil {
		err = fmt.Errorf("Error reading file for MIME type detection >> %w", err)
		return
	}

	mimeType = http.DetectContentType(sniffBuffer[:n])

	exts, _ := mime.ExtensionsByType(mimeType)
	if len(exts) > 0 {
		ext = exts[0][1:] // Remove the leading dot
	}

	if ext == "" {
		err = fmt.Errorf("Could not determine file extension for MIME type: %s", mimeType)
		return
	}

	if ext == "jpe" {
		ext = "jpeg"
	}

	_, err = file.Seek(0, io.SeekStart)
	if err != nil {
		err = fmt.Errorf("Error resetting file pointer >> %w", err)
		return
	}

	imgCfg, _, err := image.DecodeConfig(file)
	if err != nil {
		err = fmt.Errorf("Error decoding image config >> %w", err)
		return
	}

	if fileHeader.Size <= 0 {
		newSize, sizeErr := file.Seek(0, io.SeekEnd)
		if sizeErr != nil {
			err = fmt.Errorf("Error determining file size >> %w", sizeErr)
			return
		}
		fileHeader.Size = newSize
	}

	_, err = file.Seek(0, io.SeekStart)
	if err != nil {
		err = fmt.Errorf("Error resetting file pointer >> %w", err)
		return
	}

	return mimeType, ext, fileHeader.Size, imgCfg.Width, imgCfg.Height, nil
}

func createThumbnail(imageID uuid.UUID, imageExt string, db db.DB) (err error) {
	imagePath := fmt.Sprintf("./%s/%s.%s", FullSizeStoragePath, imageID, imageExt)
	in, err := os.Open(imagePath)
	if err != nil {
		err = fmt.Errorf("Error opening image file >> %w", err)
		return
	}
	defer in.Close()

	srcImg, format, err := image.Decode(in)
	if err != nil {
		err = fmt.Errorf("Error decoding image >> %w", err)
		return
	}

	bounds := srcImg.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	scale := 1.0
	if width > MaxThumbnailSize || height > MaxThumbnailSize {
		if width > height {
			scale = float64(MaxThumbnailSize) / float64(width)
		} else {
			scale = float64(MaxThumbnailSize) / float64(height)
		}
	}

	newW := int(float64(width) * scale)
	newH := int(float64(height) * scale)

	thumbImg := image.NewRGBA(image.Rect(0, 0, newW, newH))
	draw.CatmullRom.Scale(thumbImg, thumbImg.Bounds(), srcImg, bounds, draw.Over, nil)

	image := model.InsertThumbnailImage{
		ParentID:    imageID,
		VariantName: "thumb",
		StoragePath: ThumbnailStoragePath,
		Width:       newW,
		Height:      newH,
	}

	id, err := db.InsertThumbnailImage(image)

	dstPath := fmt.Sprintf("./%s", ThumbnailStoragePath)
	err = createPathIfNotExists(dstPath)
	if err != nil {
		err = fmt.Errorf("error creating thumbnail directory >> %w", err)
		return
	}

	storageFileName := fmt.Sprintf("%s.%s", id, imageExt)
	storagePath := filepath.Join(dstPath, storageFileName)

	out, err := os.Create(storagePath)
	if err != nil {
		err = fmt.Errorf("error creating thumbnail file on disk >> %w", err)
		return
	}
	defer out.Close()

	switch format {
	case "jpeg":
		err = jpeg.Encode(out, thumbImg, nil)
	case "png":
		err = png.Encode(out, thumbImg)
	case "jpg":
		err = jpeg.Encode(out, thumbImg, nil)
	default:
		err = fmt.Errorf("unsupported image format: %s", format)
	}
	if err != nil {
		err = fmt.Errorf("error encoding thumbnail image >> %w", err)
		return
	}

	return
}

func createPathIfNotExists(path string) (err error) {
	err = os.MkdirAll(path, 0o755)
	if err != nil {
		err = fmt.Errorf("error creating upload directory >> %w", err)
		return
	}

	return
}
