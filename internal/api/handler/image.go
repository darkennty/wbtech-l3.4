package handler

import (
	"errors"
	"mime"
	"net/http"
	"path/filepath"
	"strings"

	"WBTech_L3.4/internal/repository"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func (h *Handler) handleUpload(c *gin.Context) {
	file, err := c.FormFile("image")
	if err != nil {
		ReturnErrorResponse(c, h.logger, http.StatusBadRequest, "missing image file")
		return
	}

	ext := strings.ToLower(filepath.Ext(file.Filename))
	contentType := file.Header.Get("Content-Type")

	isAllowedExt := ext == ".jpg" || ext == ".jpeg" || ext == ".png"
	isAllowedType := contentType == "image/jpeg" || contentType == "image/png"

	if !isAllowedExt || !isAllowedType {
		ReturnErrorResponse(c, h.logger, http.StatusBadRequest, "only JPG, JPEG and PNG formats are allowed")
		return
	}

	if file.Size > 10<<20 {
		ReturnErrorResponse(c, h.logger, http.StatusBadRequest, "file too large (max 10MB)")
		return
	}

	f, err := file.Open()
	if err != nil {
		ReturnErrorResponse(c, h.logger, http.StatusInternalServerError, "failed to open file")
		return
	}
	defer f.Close()

	id, err := h.services.Upload(c, file.Filename, f)
	if err != nil {
		ReturnErrorResponse(c, h.logger, http.StatusInternalServerError, err.Error())
		return
	}

	ReturnResultResponse(c, http.StatusOK, gin.H{"id": id})
}

func (h *Handler) handleGetImage(c *gin.Context) {
	id := c.Param("id")
	imgType := c.Query("type")

	if _, err := uuid.Parse(id); err != nil {
		ReturnErrorResponse(c, h.logger, http.StatusBadRequest, "invalid id")
		return
	}

	img, data, err := h.services.GetImage(c, id, imgType)

	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			ReturnErrorResponse(c, h.logger, http.StatusNotFound, "image not found")
		} else {
			ReturnErrorResponse(c, h.logger, http.StatusInternalServerError, err.Error())
		}
		return
	}

	if img.Status != "completed" {
		ReturnResultResponse(c, http.StatusAccepted, gin.H{
			"id":     img.ID,
			"status": img.Status,
		})
		return
	}

	if len(data) == 0 {
		ReturnResultResponse(c, http.StatusOK, gin.H{"status": "completed"})
		return
	}

	var finalFilename string
	switch c.Query("type") {
	case "original":
		finalFilename = img.OriginalPath
	default:
		finalFilename = img.WatermarkPath
	}

	contentType := mime.TypeByExtension(filepath.Ext(finalFilename))
	if contentType == "" {
		contentType = "image/jpeg"
	}

	c.Data(http.StatusOK, contentType, data)
}

func (h *Handler) handleDeleteImage(c *gin.Context) {
	id := c.Param("id")
	if _, err := uuid.Parse(id); err != nil {
		ReturnErrorResponse(c, h.logger, http.StatusBadRequest, "invalid id")
		return
	}

	if err := h.services.Delete(c, id); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			ReturnErrorResponse(c, h.logger, http.StatusNotFound, "image not found")
		} else {
			ReturnErrorResponse(c, h.logger, http.StatusInternalServerError, err.Error())
		}
		return
	}

	ReturnResultResponse(c, http.StatusOK, gin.H{"status": "ok"})
}
