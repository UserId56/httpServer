package controllers

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/UserId56/httpServer/core/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type FileController struct {
	DB *gorm.DB
}

func NewFileController(db *gorm.DB) *FileController {
	return &FileController{
		DB: db,
	}
}

var maxUploadSize int64 = 20 * 1024 * 1024 // 10 MB

var deniedExt = map[string]bool{
	".exe": true, ".bat": true, ".cmd": true, ".com": true, ".msi": true, ".scr": true, ".dll": true, ".jar": true,
	".sh": true, ".ps1": true, ".vbs": true, ".py": true, ".pl": true, ".rb": true, ".cgi": true,
	".php": true, ".php3": true, ".phtml": true, ".asp": true, ".aspx": true, ".jsp": true, ".war": true,
	".html": true, ".htm": true, ".xhtml": true, ".svg": true,
	//".zip": true, ".rar": true, ".7z": true, ".tar": true, ".gz": true, ".bz2": true,
	//".docm": true, ".xlsm": true, ".pptm": true,
}

func (fc *FileController) FileUpload(c *gin.Context) {
	// Получаем файл из формы (поле "file")
	fileHeader, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Файл не предоставлен"})
		return
	}

	if fileHeader.Size > maxUploadSize {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Файл слишком большой"})
		return
	}

	ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
	if deniedExt[ext] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Не подерживаемый формат файла"})
		return
	}

	// Открываем источник
	src, err := fileHeader.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка на сервере"})
		return
	}
	defer src.Close()

	// Создаём папку uploads, если нет
	dstDir := "uploads"
	if err := os.MkdirAll(dstDir, 0o755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка на сервере"})
		return
	}

	// Уникальное имя файла
	id := uuid.New()
	filename := id.String() + ext
	dstPath := filepath.Join(dstDir, filename)

	// Сохраняем файл на диск
	dst, err := os.Create(dstPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot create destination file"})
		return
	}
	defer dst.Close()

	if _, err := io.Copy(dst, io.LimitReader(src, maxUploadSize+1)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot save file"})
		return
	}

	var file models.File
	file.FileId = id
	file.Name = fileHeader.Filename
	file.Path = dstPath
	userId, exist := c.Get("user_id")
	if !exist {
		c.Status(http.StatusUnauthorized)
	}
	file.OwnerID = uint(userId.(float64))
	file.FileSize = fileHeader.Size

	if err := fc.DB.Model(&models.File{}).Create(&file).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка на сервере"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id": file.FileId,
	})
}

func (fc *FileController) FileGetByID(c *gin.Context) {
	fileId := c.Param("id")
	uid, err := uuid.Parse(fileId)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Не валидный ID файла"})
		return
	}

	var file models.File
	if err := fc.DB.Where("file_id = ?", uid).First(&file).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Файл не найден"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка на сервере"})
		return
	}

	c.FileAttachment(file.Path, file.Name)
}

func (fc *FileController) FileDeleteByID(c *gin.Context) {
	fileId := c.Param("id")
	uid, err := uuid.Parse(fileId)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Не валидный ID файла"})
		return
	}

	var file models.File
	if err := fc.DB.Where("file_id = ?", uid).First(&file).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Файл не найден"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка на сервере"})
		return
	}

	if err := os.Remove(file.Path); err != nil && !os.IsNotExist(err) {
		fmt.Printf("Ошибка при удалении файла с диска: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при удалении файла с диска"})
		return
	}

	if err := fc.DB.Unscoped().Delete(&file).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка на сервере"})
		return
	}

	c.Status(http.StatusOK)
}
