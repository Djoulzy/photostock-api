package model

import (
	"time"
)

type Gallery struct {
	ID          uint64    `json:"id" gorm:"primaryKey"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
	SourceName  string    `json:"sourceName"`
	DisplayName string    `json:"displayName"`
	Copyright   string    `json:"copyRight"`
	NBItems     int       `json:"nbItems"`
	Tags        string    `json:"tags"`
	Quality     int       `json:"quality"`
	Views       int       `json:"views"`
	Rating      int       `json:"rating"`
	ThumbID     uint64    `json:"thumbId"`
	Images      []*Photo
	Thumb       *Photo
}

type Photo struct {
	ID        uint64    `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	GalleryID uint64    `json:"galleryId"`
	Hash      string    `json:"hash" gorm:"unique"`
	Ext       string    `json:"ext"`
	Height    int       `json:"height"`
	Width     int       `json:"width"`
	Quality   int       `json:"quality"`
	Rank      int       `json:"rank"`
	Full      string
}

type InfoFile struct {
	ID          uint64    `json:"id" gorm:"primaryKey"`
	SourceName  string    `json:"sourceName"`
	DisplayName string    `json:"modelName"`
	CreatedAt   time.Time `json:"createdAt"`
	Copyright   string    `json:"copyRight"`
	Tags        string    `json:"tags"`
	Rating      int       `json:"rating"`
	Images      []string  `json:"images"`
}

// Settings struct added for application settings.
// It stores a password and auditing timestamps.
type Settings struct {
	ID                uint64 `json:"id" gorm:"primaryKey"`
	AppName           string `json:"appName"`
	GalleryScreenSize string `json:"galleryScreenSize"`
	GalleryScreenCols string `json:"galleryScreenCols"`
	ContentScreenSize string `json:"contentScreenSize"`
	ContentScreenCols string `json:"contentScreenCols"`
	Password          string `json:"-"`
}

// type GalleryOut struct {
// 	Total int        `json:"total"`
// 	Data  []*Gallery `json:"data"`
// }

// type PhotoOut struct {
// 	Data []*Photo `json:"data"`
// }
