/**
 * 功能：upload.go
 * 创建时间：2026-04-18
 * 创建人：wangcw
 */

package handler

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redgreat/teweicun/internal/dto/response"
	"github.com/redgreat/teweicun/internal/pkg/errcode"
	"github.com/redgreat/teweicun/pkg/storage"
)

const (
	MaxUploadSize = 50 * 1024 * 1024 // 50MB
)

var allowedExtensions = map[string]bool{
	".jpg":  true,
	".jpeg": true,
	".png":  true,
	".pdf":  true,
	".doc":  true,
	".docx": true,
	".xls":  true,
	".xlsx": true,
}

// UploadFile handles file upload to the configured storage
func UploadFile(c *gin.Context) {
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		response.Error(c, errcode.NewAppError(errcode.ErrInvalidParam.Code, "No file uploaded", errcode.ErrInvalidParam.HTTPCode))
		return
	}
	defer file.Close()

	// Check size
	if header.Size > MaxUploadSize {
		response.Error(c, errcode.NewAppError(errcode.ErrInvalidParam.Code, "File too large (max 50MB)", errcode.ErrInvalidParam.HTTPCode))
		return
	}

	// Check extension
	ext := filepath.Ext(header.Filename)
	if !allowedExtensions[ext] {
		response.Error(c, errcode.NewAppError(errcode.ErrInvalidParam.Code, "Extension not allowed", errcode.ErrInvalidParam.HTTPCode))
		return
	}

	// Generate unique filename: year/month/day/unix_nano + ext
	now := time.Now()
	uniqueName := fmt.Sprintf("%d/%02d/%02d/%d%s", 
		now.Year(), now.Month(), now.Day(), now.UnixNano(), ext)

	// Upload
	path, err := storage.GlobalStorage.Upload(uniqueName, file)
	if err != nil {
		response.Error(c, errcode.NewAppError(errcode.ErrInternalServer.Code, "Upload failed: "+err.Error(), errcode.ErrInternalServer.HTTPCode))
		return
	}

	url, _ := storage.GlobalStorage.GetURL(path)

	response.Success(c, gin.H{
		"path": path,
		"url":  url,
		"name": header.Filename,
	})
}
