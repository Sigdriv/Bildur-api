package handler

import (
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"

	"github.com/Sigdriv/Bildur-api/model"
	"github.com/gin-gonic/gin"
)

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

	sniffSize := 512
	sniffBuffer := make([]byte, sniffSize)
	n, err := file.Read(sniffBuffer)
	if err != nil {
		log.Errorf("Error reading file for MIME type detection >> %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read file"})
		return
	}

	mimeType := http.DetectContentType(sniffBuffer[:n])
	log.Infof("Detected MIME type: %s", mimeType)

	exts, _ := mime.ExtensionsByType(mimeType)
	var ext string
	if len(exts) > 0 {
		ext = exts[0][1:] // Remove the leading dot
	}

	log.Infof("Determined file extension: %s", ext)

	if ext == "" {
		log.Warnf("Could not determine file extension for MIME type: %s", mimeType)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unsupported file type"})
		return
	}

	_, err = file.Seek(0, io.SeekStart)
	if err != nil {
		log.Errorf("Error resetting file pointer >> %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process file"})
		return
	}

	imgCfg, _, err := image.DecodeConfig(file)
	if err != nil {
		log.Errorf("Error decoding image config >> %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to decode image"})
		return
	}

	log.Infof("Image dimensions: %dx%d", imgCfg.Width, imgCfg.Height)

	if fileHeader.Size <= 0 {
		size, sizeErr := file.Seek(0, io.SeekEnd)
		if sizeErr != nil {
			log.Errorf("Error determining file size >> %v", sizeErr)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to determine file size"})
			return
		}
		fileHeader.Size = size
	}

	_, err = file.Seek(0, io.SeekStart)
	if err != nil {
		log.Errorf("Error resetting file pointer >> %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process file"})
		return
	}

	image := model.InsertImage{
		Name:        fileHeader.Filename,
		MimeType:    mimeType,
		Extension:   ext,
		Bytes:       fileHeader.Size,
		StoragePath: "media/fullsize",
		Width:       imgCfg.Width,
		Height:      imgCfg.Height,
	}

	id, err := srv.DB.InsertImage(image)
	if err != nil {
		log.Errorf("Error inserting image into database >> %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save image"})
		return
	}

	log.Infof("Successfully uploaded image with ID: %s", id)

	uploadDir := "./media/fullsize"
	err = os.MkdirAll(uploadDir, 0o755)
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

	log.Infof("File saved successfully at: %s", storagePath)

	c.JSON(http.StatusOK, gin.H{"id": id})
}
