package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/Djoulzy/IStock2/database"
	"github.com/Djoulzy/IStock2/model"
	"github.com/Djoulzy/IStock2/utils"
	"github.com/Djoulzy/Tools/confload"
)

var (
	conf = &model.ConfigData{}
	DB   database.Node
)

func main() {
	confload.Load("IS2.ini", conf)
	DB.Connect(conf)

	fileInfos, err := utils.SearchForJSONFiles(conf.AbsoluteBankPath)
	if err != nil {
		log.Fatalf("Error searching for infos.json: %v", err)
	}

	if len(fileInfos) > 0 {
		for _, fileInfo := range fileInfos {
			fmt.Printf("Found infos.json at: %s\n", fileInfo.Path)
			info, err := readAndDecodeJSON(fileInfo.Path)
			if err != nil {
				log.Printf("Error reading or decoding infos.json at %s: %v", fileInfo.Path, err)
				continue
			}
			// fmt.Printf("Decoded info: %+v\n", info)
			gal := AddGallery(info)
			thumb := AddPhotos(info)
			gal.ThumbID = thumb.ID
			DB.Db.Save(&gal)
		}
		if conf.DBDriver == "postgres" {

			setSequence := "SELECT setval('galleries_id_seq', (select max(id) from galleries), true)"
			tx := DB.Db.Exec(setSequence)
			if tx.Error != nil {
				log.Fatalf("Can't initialize sequence: %v", tx.Error)
			}
		}
	} else {
		fmt.Println("infos.json not found.")
	}
}

func AddGallery(info *model.InfoFile) *model.Gallery {
	// ext := filepath.Ext(info.Images[1])
	// hash := filepath.Base(info.Images[1][:len(info.Images[1])-len(ext)])

	newGal := model.Gallery{
		ID:          info.ID,
		SourceName:  info.SourceName,
		DisplayName: info.DisplayName,
		Copyright:   info.Copyright,
		Tags:        info.Tags,
		Rating:      info.Rating,
		CreatedAt:   info.CreatedAt,
		NBItems:     len(info.Images),
		Views:       0,
	}
	DB.Db.Create(&newGal)
	return &newGal
}

func AddPhotos(info *model.InfoFile) (thumb *model.Photo) {
	for index, image := range info.Images {
		ext := filepath.Ext(image)
		hash := filepath.Base(image[:len(image)-len(ext)])
		newPhoto := model.Photo{
			GalleryID: info.ID,
			CreatedAt: time.Now(),
			Hash:      hash,
			Ext:       ext,
			Rank:      index,
		}
		DB.Db.Create(&newPhoto)
		if index == 0 {
			thumb = &newPhoto
		}
	}
	return thumb
}

func readAndDecodeJSON(filePath string) (*model.InfoFile, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var info model.InfoFile
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&info); err != nil {
		return nil, err
	}

	return &info, nil
}
