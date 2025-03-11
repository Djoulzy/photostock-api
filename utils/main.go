package utils

import (
	"crypto/md5"
	"encoding/hex"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type FileInfo struct {
	Path string
}

const (
	CONTENT_DIR = "/content"
	THUMB_DIR   = "/thumbs"
)

func GetGalleryPath(stockBaseDir string, galId uint64) string {
	base := strings.TrimRight(stockBaseDir, "/")
	list := strings.Split(strconv.FormatUint(uint64(galId), 10), "")
	return base + "/" + strings.Join(list, "/") + CONTENT_DIR
}

func GetThumbPath(galId uint64, thumb string) string {
	list := strings.Split(strconv.FormatUint(uint64(galId), 10), "")
	return "img/" + strings.Join(list, "/") + CONTENT_DIR + THUMB_DIR + "/" + thumb + ".png"
}

func GetFullImagePath(galId uint64, hash string, ext string) string {
	list := strings.Split(strconv.FormatUint(uint64(galId), 10), "")
	return "img/" + strings.Join(list, "/") + CONTENT_DIR + "/" + hash + ext
}

func SearchForJSONFiles(rootDir string) ([]FileInfo, error) {
	var fileInfos []FileInfo

	err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && info.Name() == "infos.json" {
			fileInfos = append(fileInfos, FileInfo{Path: path})
		}
		return nil
	})

	return fileInfos, err
}

func SearchForThumbDir(rootDir string) ([]FileInfo, error) {
	var fileInfos []FileInfo

	err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() && info.Name() == "thumbs" {
			fileInfos = append(fileInfos, FileInfo{Path: path})
		}
		return nil
	})

	return fileInfos, err
}

// MD5Hash returns the MD5 hash of the input string
func MD5Hash(text string) string {
	hash := md5.Sum([]byte(text))
	return hex.EncodeToString(hash[:])
}
