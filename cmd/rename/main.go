package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/Djoulzy/Tools/confload"
	"github.com/Djoulzy/photostock-api/database"
	"github.com/Djoulzy/photostock-api/model"
)

var (
	conf = &model.ConfigData{}
	DB   database.Node
)

func main() {
	confload.Load("IS2.ini", conf)
	DB.Connect(conf)

	names, err := LoadCSVFile("assets/name.csv")
	if err != nil {
		fmt.Printf("Erreur lors du chargement du CSV : %v\n", err)
		return
	}
	fmt.Println("Names loaded...")

	copyrights, err := LoadCSVFile("assets/copyright.csv")
	if err != nil {
		fmt.Printf("Erreur lors du chargement du CSV : %v\n", err)
		return
	}
	fmt.Println("Copyright loaded...")

	updateGalleries(names, copyrights)
}

func findToken(source string, tokens []string) (string, bool) {
	for _, token := range tokens {
		if strings.Contains(source, token) {
			return token, true
		}
	}
	return "", false
}

func updateGalleries(names []string, copyrights []string) (err error) {
	var list []*model.Gallery
	var needUpdate bool = false

	result := DB.Db.Table("galleries").Find(&list)
	if result.Error != nil || result.RowsAffected == 0 {
		// clog.Debug("pgsql", "GetListByTag", "No data found: %v", result.Error)
		log.Printf("No data found: %v\n", result.Error)
		return result.Error
	}

	for _, gal := range list {
		needUpdate = false
		fmt.Println(gal.SourceName)
		if token, found := findToken(gal.SourceName, names); found {
			gal.DisplayName = token
			needUpdate = true
		}
		if token, found := findToken(gal.SourceName, copyrights); found {
			gal.Copyright = token
			needUpdate = true
		}
		if needUpdate {
			result := DB.Db.Save(gal)
			if result.Error != nil || result.RowsAffected == 0 {
				return result.Error
			}
		}
	}
	return nil
}

// LoadCSVFile lit un fichier CSV dont la première ligne est l'en-tête "label" et retourne un slice de chaînes.
func LoadCSVFile(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	// Vérifier que le fichier comporte au moins une ligne d'en-tête
	if len(records) < 1 {
		return nil, fmt.Errorf("fichier CSV vide")
	}

	// Vérifier l'en-tête
	if len(records[0]) != 1 || records[0][0] != "label" {
		return nil, fmt.Errorf("en-tête invalide, attendu: 'label'")
	}

	// Parcourir les lignes suivantes pour récupérer les labels.
	var labels []string
	for _, row := range records[1:] {
		if len(row) > 0 {
			labels = append(labels, row[0])
		}
	}
	return labels, nil
}
