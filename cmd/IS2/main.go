package main

import (
	"log"

	"github.com/Djoulzy/Tools/confload"
	"github.com/Djoulzy/photostock-api/database"
	docs "github.com/Djoulzy/photostock-api/docs"
	"github.com/Djoulzy/photostock-api/flow"
	"github.com/Djoulzy/photostock-api/model"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

var (
	conf = &model.ConfigData{}
	DB   database.Node
)

func main() {
	confload.Load("IS2.ini", conf)
	DB.Connect(conf)
	flow.StartUploadFileAssembler(&DB, conf)

	log.Printf("\n\nAPI DOCS: http://%s/swagger/index.html\n\n", conf.Globals.HTTP_addr)

	if conf.Globals.Mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}
	router := gin.Default()

	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"*"}
	router.Use(cors.New(config))

	docs.SwaggerInfo.BasePath = "/api/v1"
	router.Static("/img", "./img")

	v1 := router.Group(docs.SwaggerInfo.BasePath)
	{
		v1.OPTIONS("/*any", options)
		v1.GET("/gallery", getGallery)
		v1.GET("/gallery/:id", getGalleryByID)
		v1.GET("/gallery/:id/update-views", updateGalleryViewsByID)
		v1.PUT("/gallery/:id", putGalleryByID)
		v1.DELETE("/gallery/:id", deleteGalleryByID)
		v1.POST("/gallery/mix", mixGalleries)

		v1.GET("/photo", getPhotoByGalleryID)
		v1.GET("/photo/:id", getPhoto)
		v1.GET("/thumb/:galId/:imgId/:hash/:size", getThumbByID)

		v1.GET("/settings", getSettings)
		v1.PUT("/settings", updateSettings)
		v1.POST("/settings/auth", comparePassword)

		v1.GET("/import", getImport)

		// v1.OPTIONS("/upload", flow.CorsHandler)
		v1.GET("/upload", getUpload)
		v1.POST("/upload", postUpload)
	}

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))
	router.Run(conf.Globals.HTTP_addr)
}
