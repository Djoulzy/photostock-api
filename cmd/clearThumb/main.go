package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/Djoulzy/IStock2/model"
	"github.com/Djoulzy/IStock2/utils"
	"github.com/Djoulzy/Tools/confload"
)

var (
	conf = &model.ConfigData{}
)

func main() {
	confload.Load("IS2.ini", conf)

	fileInfos, err := utils.SearchForThumbDir(conf.AbsoluteBankPath)
	if err != nil {
		log.Fatalf("Error searching for thumbs dir: %v", err)
	}

	if len(fileInfos) > 0 {
		for _, fileInfo := range fileInfos {
			fmt.Printf("thumbs dir found at: %s\n", fileInfo.Path)
			ClearFilesInDir(fileInfo.Path)
		}
	} else {
		fmt.Println("thumbs dir not found.")
	}
}

// ClearFilesInDir efface tous les fichiers (mais pas les sous-répertoires) dans le répertoire donné.
func ClearFilesInDir(dirPath string) error {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			err := os.Remove(filepath.Join(dirPath, entry.Name()))
			fmt.Printf("removing: %s\n", filepath.Join(dirPath, entry.Name()))
			if err != nil {
				return err
			}
		}
	}

	return nil
}
