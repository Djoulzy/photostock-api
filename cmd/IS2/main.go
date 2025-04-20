package main

import (
	"log"
	"net"
	"os"
	"strings"

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

func GetOutboundIP() net.IP {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP
}

func main() {
	var serverAddress string
	confload.Load("IS2.ini", conf)
	DB.Connect(conf)
	flow.StartUploadFileAssembler(&DB, conf)

	if conf.Globals.HTTP_addr == "" {
		serverAddress = GetOutboundIP().String()
	} else {
		serverAddress = conf.Globals.HTTP_addr
	}

	if conf.Globals.HTTP_port == "" {
		serverAddress = serverAddress + ":" + os.Getenv("PHOTOSTOCK_API_PORT")
	} else {
		serverAddress = serverAddress + ":" + conf.Globals.HTTP_port
	}

	log.Printf("\n\nAPI DOCS: http://%s/swagger/index.html\n\n", serverAddress)

	if conf.Globals.Mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}
	router := gin.Default()

	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"*"}
	router.Use(cors.New(config))

	docs.SwaggerInfo.BasePath = "/api/v1"
	base := strings.TrimRight(conf.AbsoluteBankPath, "/")
	router.Static("/img", base)
	log.Println("Serving images from: " + base)

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
		v1.PATCH("/settings/auth", changePassword)

		v1.GET("/import", getImport)

		// v1.OPTIONS("/upload", flow.CorsHandler)
		v1.GET("/upload", getUpload)
		v1.POST("/upload", postUpload)
	}

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))
	router.Run(serverAddress)
}
