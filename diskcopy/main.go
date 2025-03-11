package diskcopy

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"io"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/Djoulzy/IStock2/database"
	"github.com/Djoulzy/IStock2/model"
	"github.com/Djoulzy/IStock2/utils"
	"github.com/disintegration/imaging"
)

func CreateStockDir(stockBaseDir string, galId uint64) (string, error) {
	contentDir := utils.GetGalleryPath(stockBaseDir, galId)
	thumbDir := contentDir + utils.THUMB_DIR

	// TODO: Check if dir is empty
	if _, err := os.Stat(thumbDir); os.IsNotExist(err) {
		os.MkdirAll(thumbDir, os.ModePerm)
	} else {
		return "", errors.New("directory already exists")
	}

	return contentDir, nil
}

func CreateUniqHash(DB *database.Node, fileName string, to string, ext string) (string, error) {
	hash := sha1.New()

	bytesRead, err := os.ReadFile(fileName)
	if err != nil {
		return "", err
	}
	hash.Write(bytesRead)

	sum := hex.EncodeToString(hash.Sum(nil))
	if DB.IsHashExists(sum) {
		log.Printf("%s Exists !! -> %s\n", fileName, sum)
		return "", errors.New("photo exists in DB")
	}

	err = os.WriteFile(to+"/"+sum+ext, bytesRead, 0755)
	if err != nil {
		return "", err
	}

	return sum, nil
}

func CreateThumb(sourcePath string, width int, destPath string) (err error) {
	var thumb image.Image

	source, err := imaging.Open(sourcePath)
	if err != nil {
		log.Println(err)
		return err
	}

	// srcWidth := source.Bounds().Max.X
	// srcHeight := source.Bounds().Max.Y

	// if srcWidth >= srcHeight {
	thumb = imaging.Resize(source, width, 0, imaging.Lanczos)
	// }else {
	// 	thumb = imaging.Resize(source, 0, width, imaging.Lanczos)
	// }
	err = imaging.Save(thumb, destPath)
	if err != nil {
		return err
	}
	return nil
}

func CopyFiles(DB *database.Node, galId uint64, from string, to string) (int, uint64, error) {
	var thumbId uint64
	files, err := os.ReadDir(from)
	if err != nil {
		log.Fatal(err)
		return 0, 0, err
	}

	sort.Slice(files, func(i, j int) bool {
		fileI, _ := files[i].Info()
		fileJ, _ := files[j].Info()
		return fileI.ModTime().Before(fileJ.ModTime())
	})

	count := 0
	for index, f := range files {
		fileName := f.Name()
		ext := strings.ToLower(filepath.Ext(fileName))
		if f.IsDir() || (ext != ".jpg" && ext != ".jpeg" && ext != ".png") {
			continue
		}

		srcFile := from + "/" + fileName
		sum, err := CreateUniqHash(DB, srcFile, to, ext)
		if err != nil {
			log.Println(err)
			continue
		}

		imageData, err := imaging.Open(srcFile)
		if err != nil {
			log.Println(err)
			continue
		}

		newPhoto := model.Photo{
			GalleryID: galId,
			Hash:      sum,
			Ext:       ext,
			Width:     imageData.Bounds().Max.X,
			Height:    imageData.Bounds().Max.Y,
			Rank:      count,
		}

		// CreateThumb(imageData, 150, to+utils.THUMB_DIR+"/"+sum+".png")

		DB.Db.Create(&newPhoto)
		// if preview == "" {
		// 	preview = sum
		// }
		if index == 1 {
			thumbId = newPhoto.ID
		}
		count++
	}
	return count, thumbId, err
}

func WriteInfoFile(DB *database.Node, stockBaseDir string, galId uint64) error {
	path := utils.GetGalleryPath(stockBaseDir, galId)
	infos, _, err := DB.GetGallery(galId, "", "", 0, 0)
	if err != nil {
		return err
	}
	hash, err := DB.GetPhotosHashListByGallery(galId)
	if err != nil {
		return err
	}
	data := model.InfoFile{
		ID:          infos.ID,
		SourceName:  infos.SourceName,
		DisplayName: infos.DisplayName,
		CreatedAt:   infos.CreatedAt,
		Copyright:   infos.Copyright,
		Tags:        infos.Tags,
		Rating:      infos.Rating,
		Images:      hash,
	}
	js, err := json.Marshal(data)
	if err != nil {
		return err
	}
	err = os.WriteFile(path+"/infos.json", js, 0644)
	if err != nil {
		return err
	}
	return nil
}

func ImportSelected(DB *database.Node, importBaseDir string, name string, stockBaseDir string) error {
	newGal := model.Gallery{
		SourceName:  name,
		DisplayName: name,
		ThumbID:     1,
		Tags:        "NEW",
		Views:       0,
	}
	DB.Db.Create(&newGal)

	destDir, err := CreateStockDir(stockBaseDir, newGal.ID)
	if err != nil {
		DB.Db.Delete(&newGal)
		return err
	}
	nb, thumbId, _ := CopyFiles(DB, newGal.ID, importBaseDir+"/"+name, destDir)
	if nb == 0 {
		DB.Db.Delete(&newGal)
		return errors.New("no new photos to imports")
	}

	newGal.NBItems = nb
	newGal.ThumbID = thumbId
	DB.Db.Save(&newGal)
	err = WriteInfoFile(DB, stockBaseDir, newGal.ID)

	return err
}

func LookupNewDir(DB *database.Node, importBaseDir string, stockBaseDir string) error {
	files, err := os.ReadDir(importBaseDir)
	if err != nil {
		return err
	}

	for _, f := range files {
		fileName := f.Name()
		if strings.HasPrefix(fileName, ".") || strings.HasPrefix(fileName, "@") || strings.HasPrefix(fileName, "_") || strings.HasPrefix(fileName, "thumbs") || (filepath.Ext(fileName) == ".part") {
			continue
		}
		if f.IsDir() {
			fmt.Printf("Importing: %s\n", fileName)
			err = ImportSelected(DB, importBaseDir, fileName, stockBaseDir)
			if err != nil {
				log.Print(err)
				return err
			}
			os.RemoveAll(importBaseDir + "/" + fileName)
			fmt.Println(" -> OK")
		}
	}
	return nil
}

func simpleCopy(src, dst string) error {
	source, err := os.Open(src)
	if err != nil {
		return err
	}
	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destination.Close()

	_, err = io.Copy(destination, source)
	return err
}

func MovePhotos(stockBaseDir string, id_from uint64, id_to uint64) error {
	from := utils.GetGalleryPath(stockBaseDir, id_from)
	to := utils.GetGalleryPath(stockBaseDir, id_to)

	files, err := os.ReadDir(from)
	if err != nil {
		return err
	}

	for _, f := range files {
		fileName := f.Name()
		ext := strings.ToLower(filepath.Ext(fileName))
		if f.IsDir() || (ext != ".jpg" && ext != ".jpeg" && ext != ".png") {
			continue
		}
		// thumbName := strings.Replace(fileName, ext, ".png", 1)

		// Copy full size
		if err = simpleCopy(from+"/"+fileName, to+"/"+fileName); err != nil {
			log.Printf("Copy: %s to %s: ERR", from+"/"+fileName, to+"/"+fileName)
			return err
		}
		log.Printf("Copy: %s to %s: OK", from+"/"+fileName, to+"/"+fileName)

		// Copy thumbs
		// if err = simpleCopy(from+"/thumbs/"+thumbName, to+"/thumbs/"+thumbName); err != nil {
		// 	log.Printf("Copy: %s to %s: ERR", from+"/thumbs/"+thumbName, to+"/thumbs/"+thumbName)
		// 	return err
		// }
		// log.Printf("Copy: %s to %s: OK", from+"/thumbs/"+thumbName, to+"/thumbs/"+thumbName)
	}

	// Remove Origin Dir
	return os.RemoveAll(from)
}

func MixGalleries(DB *database.Node, stockBaseDir string, origin uint64, dest uint64) (err error) {
	if err = DB.MixGalleries(origin, dest); err != nil {
		return err
	}
	if err := MovePhotos(stockBaseDir, origin, dest); err != nil {
		return err
	}

	if err = DB.SuppressGallery(origin); err != nil {
		return err
	}
	return nil
}

func DeleteGallery(DB *database.Node, stockBaseDir string, id uint64) (err error) {
	if err := DeleteFullGallery(stockBaseDir, id); err != nil {
		return err
	}
	if err = DB.SuppressPhotosFromGallery(id); err != nil {
		return err
	}
	if err = DB.SuppressGallery(id); err != nil {
		return err
	}
	return nil
}

func DeleteFullGallery(stockBaseDir string, id uint64) error {
	path := utils.GetGalleryPath(stockBaseDir, id)
	return os.RemoveAll(path)
}
