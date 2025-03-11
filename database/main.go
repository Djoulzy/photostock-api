package database

import (
	"log"

	"github.com/glebarez/sqlite"
	_ "github.com/lib/pq"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/Djoulzy/photostock-api/model"
	"github.com/Djoulzy/photostock-api/utils"
)

type Node struct {
	Db   *gorm.DB
	Conf *model.ConfigData
}

func (n *Node) Connect(conf *model.ConfigData) {
	var err error

	n.Conf = conf
	if conf.DBDriver == "postgres" {
		connStr := "host=" + conf.Postgres.DBHost + " port=" + conf.Postgres.DBPort + " user=" + conf.Postgres.DBUser + " password=" + conf.Postgres.DBPasswd + " dbname=" + conf.Postgres.DBName + " sslmode=disable"
		n.Db, err = gorm.Open(postgres.Open(connStr), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Silent),
		})
		if err != nil {
			panic(err)
		}
		log.Println("Connected to Postgres: " + conf.Postgres.DBName + "@" + conf.Postgres.DBHost)
	} else {
		n.Db, err = gorm.Open(sqlite.Open(conf.SQLite.DBPath), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Silent),
		})
		if err != nil {
			panic(err)
		}
		log.Println("Connected to SQLite: " + conf.SQLite.DBPath)
	}

	if err = n.Db.AutoMigrate(&model.Gallery{}); err != nil {
		panic(err)
	}
	if err = n.Db.AutoMigrate(&model.Photo{}); err != nil {
		panic(err)
	}
	if err = n.Db.AutoMigrate(&model.Settings{}); err != nil {
		panic(err)
	}
	var count int64
	n.Db.Model(&model.Settings{}).Count(&count)
	if count == 0 {
		if err := n.AddDefaultSettings(); err != nil {
			panic(err)
		}
	}
}

func (n *Node) GetConnection() interface{} {
	return n.Db
}

func (n *Node) IsHashExists(hash string) bool {
	var photo model.Photo
	result := n.Db.First(&photo, "hash = ?", hash)
	if result.Error != nil || result.RowsAffected == 0 {
		return false
	}
	return true
}

func (n *Node) getQueryWithParams(sort string, order string, start int, end int, debug bool) *gorm.DB {
	query := n.Db.Session(&gorm.Session{DryRun: debug})
	// 0-12  -> offset 0  limit 12
	// 12-24 -> offset 12 limit 12
	// 24-36 -> offset 24 limit 12
	limit := end - start
	if sort == "date" {
		sort = "created_at"
	}
	if sort != "" {
		query = query.Order(sort + " " + order)
	}
	if start != 0 {
		query = query.Offset(start)
	}
	if limit != 0 {
		query = query.Limit(limit)
	}
	if debug {
		test := query.Statement
		log.Println(test.SQL.String())
	}
	log.Printf("Offset: %d - Limit: %d\n", start, limit)
	return query
}

func (n *Node) GetGalleryList(sort string, order string, start int, end int) (list []*model.Gallery, total int64, err error) {
	n.Db.Table("galleries").Count(&total)
	query := n.getQueryWithParams(sort, order, start, end, false)
	result := query.Find(&list)
	if result.Error != nil || result.RowsAffected == 0 {
		// clog.Debug("pgsql", "GetListByTag", "No data found: %v", result.Error)
		log.Printf("No data found: %v\n", result.Error)
		return nil, 0, result.Error
	}

	for _, gal := range list {
		gal.Thumb, _ = n.GetPhoto(gal.ThumbID)
	}
	// clog.Debug("pgsql", "GetListByTag", "Found %d items", result.RowsAffected)
	return list, total, nil
}

func (n *Node) GetGallery(id uint64, sort string, order string, start int, end int) (one *model.Gallery, total int64, err error) {
	result := n.Db.First(&one, "id = ?", id)
	photos, total, err := n.GetPhotosByGallery(id, sort, order, start, end)
	if err != nil {
		return nil, 0, err
	}
	one.Images = photos
	if result.Error != nil || result.RowsAffected == 0 {
		return nil, total, result.Error
	}
	return one, total, nil
}

func (n *Node) GetPhoto(id uint64) (one *model.Photo, err error) {
	result := n.Db.First(&one, "id = ?", id)
	if result.Error != nil || result.RowsAffected == 0 {
		return nil, result.Error
	}
	return one, nil
}

func (n *Node) GetPhotosByGallery(id uint64, sort string, order string, start int, end int) (list []*model.Photo, total int64, err error) {
	n.Db.Table("photos").Where("gallery_id = ?", id).Count(&total)
	query := n.getQueryWithParams(sort, order, start, end, false)

	result := query.Where("gallery_id = ?", id).Find(&list)
	if result.Error != nil || result.RowsAffected == 0 {
		log.Println(result.Error)
		return nil, 0, result.Error
	}
	for _, img := range list {
		// img.Thumb = utils.GetThumbPath(id, img.Hash)
		img.Full = utils.GetFullImagePath(id, img.Hash, img.Ext)
	}
	return list, total, nil
}

func (n *Node) GetPhotosHashListByGallery(id uint64) (hash []string, err error) {
	var list []*model.Photo
	result := n.Db.Select("hash, ext").Where("gallery_id = ?", id).Order("rank ASC").Find(&list)
	if result.Error != nil || result.RowsAffected == 0 {
		log.Println(result.Error)
		return nil, result.Error
	}
	hash = make([]string, len(list))
	for index, str := range list {
		hash[index] = str.Hash + str.Ext
	}
	return hash, nil
}

func (n *Node) UpdateGalleryViews(id uint64) error {
	gal, _, _ := n.GetGallery(id, "", "", 0, 0)
	gal.Views += 1
	result := n.Db.Save(gal)
	if result.Error != nil || result.RowsAffected == 0 {
		return result.Error
	}
	return nil
}

func (n *Node) UpdateGalleryInfos(id uint64) error {
	return nil
}

func (n *Node) MixGalleries(origin uint64, dest uint64) error {
	result := n.Db.Model(&model.Photo{}).Where("gallery_id = ?", origin).Update("gallery_id", dest)
	if result.Error != nil || result.RowsAffected == 0 {
		return result.Error
	}
	return nil
}

func (n *Node) SuppressGallery(id uint64) error {
	result := n.Db.Delete(&model.Gallery{}, id)
	if result.Error != nil || result.RowsAffected == 0 {
		return result.Error
	}
	return nil
}

func (n *Node) SuppressPhotosFromGallery(id uint64) error {
	result := n.Db.Where("gallery_id = ?", id).Delete(&model.Photo{})
	if result.Error != nil || result.RowsAffected == 0 {
		return result.Error
	}
	return nil
}

func (n *Node) AddDefaultSettings() error {
	// Update the fields below to match the actual model.Settings structure and your desired defaults.
	defaultSettings := model.Settings{
		AppName:           "PhotoStock",
		GalleryScreenSize: "600,1024,1600",
		GalleryScreenCols: "2,3,4",
		ContentScreenSize: "600,1024,1600",
		ContentScreenCols: "3,6,10",
		Password:          utils.MD5Hash("admin"), // change as needed
	}
	result := n.Db.Create(&defaultSettings)
	if result.Error != nil {
		return result.Error
	}
	return nil
}
