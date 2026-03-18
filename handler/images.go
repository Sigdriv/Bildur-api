package handler

import (
	"fmt"
	"net/http"

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
