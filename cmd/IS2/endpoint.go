package main

// @BasePath /api/v1

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/Djoulzy/photostock-api/diskcopy"
	"github.com/Djoulzy/photostock-api/flow"
	"github.com/Djoulzy/photostock-api/model"
	"github.com/Djoulzy/photostock-api/utils"
	"github.com/gin-gonic/gin"
)

func JSON(c *gin.Context, code int, size int, obj interface{}) {
	c.Header("Access-Control-Allow-Origin", "*")
	c.Header("Access-Control-Allow-Methods", "*")
	c.Header("Access-Control-Allow-Headers", "*")
	c.Header("Access-Control-Expose-Headers", "X-Total-Count")
	c.Header("X-Total-Count", strconv.Itoa(size))
	c.JSON(code, obj)
}

func options(c *gin.Context) {
	JSON(c, http.StatusOK, 0, "OK")
}

type Form struct {
	Source int `json:"srcId" binding:"required"`
	Dest   int `json:"destId" binding:"required"`
}

type PasswordForm struct {
	Password string `json:"password" binding:"required"`
}

// @Summary			List all galleries
// @Schemes
// @Description		List all galleries
// @Tags			gallery
// @Accept			json
// @Produce			json
// @Success			200 {object} []model.Gallery
// @Router			/gallery [get]
// @Param			_start query int false "Offset"
// @Param			_end query int false "Limit"
// @Param			_sort query string false "Sort by"
// @Param			_order query string false "Order"
func getGallery(c *gin.Context) {
	start, _ := strconv.Atoi(c.Query("_start"))
	end, _ := strconv.Atoi(c.Query("_end"))
	sort := c.Query("_sort")
	order := c.Query("_order")

	list, total, err := DB.GetGalleryList(sort, order, start, end)

	// response := GalleryOut{
	// 	Total: len(list),
	// 	Data:  list,
	// }

	if err != nil {
		JSON(c, http.StatusNoContent, 0, nil)
	} else {
		JSON(c, http.StatusOK, int(total), list)
	}
}

// @Summary Get one Gallery
// @Schemes
// @Description Get one gallery infos by its ID
// @Tags gallery
// @Accept json
// @Produce json
// @Success 200 {object} model.Gallery
// @Router /gallery/{id} [get]
// @Param id path int true "Gallery ID"
// @Param _start query int false "Offset"
// @Param _end query int false "Limit"
// @Param _sort query string false "Sort by"
// @Param _order query string false "Order"
func getGalleryByID(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	start, _ := strconv.Atoi(c.Query("_start"))
	end, _ := strconv.Atoi(c.Query("_end"))
	sort := c.Query("_sort")
	order := c.Query("_order")

	list, total, err := DB.GetGallery(id, sort, order, start, end)

	if err != nil {
		JSON(c, http.StatusNoContent, 0, nil)
	} else {
		JSON(c, http.StatusOK, int(total), list)
	}
}

// @Summary Import new Gallery
// @Schemes
// @Description Start import process
// @Tags import
// @Accept json
// @Produce json
// @Success 200 {string} string OK
// @Failure 204 {string} string Error
// @Router /import [get]
func getImport(c *gin.Context) {
	err := diskcopy.LookupNewDir(&DB, conf.ImportDir, conf.AbsoluteBankPath)
	if err != nil {
		c.String(http.StatusBadRequest, "%s", err)
	} else {
		JSON(c, http.StatusOK, 0, "Ok")
	}
}

// @Summary Get photo
// @Schemes
// @Description 	Get photo list from parameters
// @Tags			photo
// @Accept			json
// @Produce			json
// @Success 		200 {object} []model.Photo
// @Router			/photo [get]
// @Param			gallery_id query int true "Gallery ID"
// @Param			_start query int false "Offset"
// @Param			_end query int false "Limit"
// @Param			_sort query string false "Sort by"
// @Param			_order query string false "Order"
func getPhotoByGalleryID(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Query("gallery_id"), 10, 64)
	// hash := c.Query("hashOnly")
	start, _ := strconv.Atoi(c.Query("_start"))
	end, _ := strconv.Atoi(c.Query("_end"))
	sort := c.Query("_sort")
	order := c.Query("_order")

	// if hash != "" {
	// 	list, err := DB.GetPhotosHashListByGallery(id)
	// 	if err != nil {
	// 		JSON(c, http.StatusNoContent, 0, nil)
	// 	} else {
	// 		JSON(c, http.StatusOK, len(list), list)
	// 	}
	// } else {
	list, total, err := DB.GetPhotosByGallery(id, sort, order, start, end)
	if err != nil {
		JSON(c, http.StatusNoContent, 0, nil)
	} else {
		JSON(c, http.StatusOK, int(total), list)
	}
	// }
	// response := PhotoOut{
	// 	Data: list,
	// }

}

// @Summary Get photo
// @Schemes
// @Description Get one photo
// @Tags photo
// @Accept json
// @Produce json
// @Param   id path int true "Photo ID"
// @Success 200 {object} model.Photo
// @Router /photo/{id} [get]
func getPhoto(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	list, err := DB.GetPhoto(id)

	if err != nil {
		JSON(c, http.StatusNoContent, 0, nil)
	} else {
		JSON(c, http.StatusOK, 1, list)
	}
}

// @Summary Get thumbnail by ID
// @Schemes
// @Description Get thumbnail information by its ID
// @Tags thumb
// @Accept json
// @Param galId path int true "Gallery ID"
// @Param imgId path int true "Image ID"
// @Param hash path string true "Thumb hash"
// @Param size path int true "Thumb width in pixel"
// @Success 200 {string} ok
// @Router /thumb/{galId}/{imgId}/{hash}/{size} [get]
func getThumbByID(c *gin.Context) {
	galId, err := strconv.ParseUint(c.Param("galId"), 10, 64)
	if err != nil {
		JSON(c, http.StatusBadRequest, 0, nil)
		return
	}
	imgId, err := strconv.ParseUint(c.Param("imgId"), 10, 64)
	if err != nil {
		JSON(c, http.StatusBadRequest, 0, nil)
		return
	}
	size, err := strconv.ParseUint(c.Param("size"), 10, 64)
	if err != nil {
		JSON(c, http.StatusBadRequest, 0, nil)
		return
	}
	log.Println("OK")

	hash := c.Param("hash")
	galPath := utils.GetGalleryPath(conf.AbsoluteBankPath, galId)
	thumbPath := fmt.Sprintf("%s/thumbs/%s@%d.png", galPath, hash, size)
	log.Println(thumbPath)
	if _, err := os.Stat(thumbPath); err != nil {
		if os.IsNotExist(err) {
			photo, err := DB.GetPhoto(imgId)
			if err != nil {
				JSON(c, http.StatusBadRequest, 0, nil)
				return
			}
			sourcePath := galPath + "/" + hash + photo.Ext
			err = diskcopy.CreateThumb(sourcePath, int(size), thumbPath)
			if err != nil {
				JSON(c, http.StatusBadRequest, 0, nil)
				return
			}
		}
	}
	c.File(thumbPath)
}

// @Summary Update Gallery views
// @Schemes
// @Description Set views +1
// @Tags gallery
// @Accept json
// @Produce json
// @Param   id path int true "Gallery ID"
// @Success 200 {string} ok
// @Router /gallery/{id}/update-views [get]
func updateGalleryViewsByID(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	err := DB.UpdateGalleryViews(id)

	if err != nil {
		JSON(c, http.StatusNoContent, 0, nil)
	} else {
		JSON(c, http.StatusOK, 1, "OK")
	}
}

// @Summary Update Gallery infos
// @Schemes
// @Description Set Model Name and tags
// @Tags gallery
// @Accept json
// @Produce json
// @Param   id path int true "Gallery ID"
// @Success 200 {object} model.Gallery
// @Router /gallery/{id} [put]
func putGalleryByID(c *gin.Context) {
	var gal model.Gallery

	// id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := c.BindJSON(&gal); err != nil {
		return
	}
	log.Printf("%v", gal)
	DB.Db.Save(gal)

	// if err != nil {
	// 	JSON(c, http.StatusNoContent, 0, nil)
	// } else {
	JSON(c, http.StatusOK, 1, gal)
	// }
}

// @Summary			Delete Gallery
// @Schemes
// @Description		Delete Gallery and photos
// @Tags gallery
// @Accept json
// @Produce json
// @Param			id path int true "Gallery ID"
// @Success			200 {string} ok
// @Router			/gallery/{id} [delete]
func deleteGalleryByID(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)

	err := diskcopy.DeleteGallery(&DB, conf.AbsoluteBankPath, id)

	if err != nil {
		JSON(c, http.StatusNoContent, 0, nil)
	} else {
		JSON(c, http.StatusOK, 1, "OK")
	}
}

// @Summary			Mix two galleries in one
// @Schemes
// @Description		mix 2 gals
// @Tags			gallery
// @Accept			json
// @Produce			json
// @Param			data body Form true "Source and Destination IDs"
// @Success			200 {string} ok
// @Router			/gallery/mix [post]
func mixGalleries(c *gin.Context) {
	var form Form

	if err := c.ShouldBind(&form); err != nil {
		c.String(http.StatusBadRequest, "bad request: %v", err)
		return
	}

	log.Printf("from %d to %d", form.Source, form.Dest)

	err := diskcopy.MixGalleries(&DB, conf.AbsoluteBankPath, uint64(form.Source), uint64(form.Dest))

	if err != nil {
		JSON(c, http.StatusNoContent, 0, nil)
	} else {
		JSON(c, http.StatusOK, 1, "OK")
	}
}

// @Summary		Compare password
// @Schemes
// @Description Compare provided MD5 encoded password (in plain text, which is then hashed) with the stored settings password.
// @Tags		settings
// @Accept		json
// @Produce		json
// @Param       password body PasswordForm true "Password"
// @Success     200 {boolean} boolean "true if the password matches, false otherwise"
// @Router      /settings/auth [post]
func comparePassword(c *gin.Context) {
	var form PasswordForm
	if err := c.ShouldBindJSON(&form); err != nil {
		JSON(c, http.StatusBadRequest, 0, nil)
		return
	}

	// Retrieve the stored settings. Assuming there's one settings record.
	var settings model.Settings
	if err := DB.Db.First(&settings).Error; err != nil {
		JSON(c, http.StatusNotFound, 0, nil)
		return
	}
	log.Printf("form: %s - pass: %s", form.Password, settings.Password)

	// Compare the computed hash with the stored password.
	valid := form.Password == settings.Password
	JSON(c, http.StatusOK, 0, valid)
}

// @Summary		change password
// @Schemes
// @Description Update the admin password.
// @Tags		settings
// @Accept		json
// @Produce		json
// @Param       password body PasswordForm true "Password"
// @Success     200 {boolean} boolean "true if the password is correctly updated, false otherwise"
// @Router      /settings/auth [patch]
func changePassword(c *gin.Context) {
	var form PasswordForm
	if err := c.ShouldBindJSON(&form); err != nil {
		JSON(c, http.StatusBadRequest, 0, nil)
		return
	}

	// Retrieve the stored settings. Assuming there's one settings record.
	var settings model.Settings
	if err := DB.Db.First(&settings).Error; err != nil {
		JSON(c, http.StatusNotFound, 0, nil)
		return
	}
	settings.Password = form.Password
	if err := DB.Db.Save(&settings).Error; err != nil {
		JSON(c, http.StatusInternalServerError, 0, err.Error())
		return
	}
	JSON(c, http.StatusOK, 0, "ok")
}

// @Summary Get Settings
// @Schemes
// @Description Returns the settings stored in the database.
// @Tags settings
// @Accept json
// @Produce json
// @Success 200 {object} model.Settings
// @Router /settings [get]
func getSettings(c *gin.Context) {
	var settings model.Settings
	if err := DB.Db.First(&settings).Error; err != nil {
		JSON(c, http.StatusNotFound, 0, nil)
		return
	}
	JSON(c, http.StatusOK, 1, settings)
}

// @Summary Update Settings
// @Schemes
// @Description Updates the settings stored in the database.
// @Tags settings
// @Accept json
// @Produce json
// @Param settings body model.Settings true "Settings object"
// @Success 200 {object} model.Settings
// @Router /settings [put]
func updateSettings(c *gin.Context) {
	var newSettings model.Settings
	if err := c.BindJSON(&newSettings); err != nil {
		JSON(c, http.StatusBadRequest, 0, fmt.Sprintf("bad request: %v", err))
		return
	}

	// Retrieve the current settings record. Assuming there is only one settings record.
	var settings model.Settings
	if err := DB.Db.First(&settings).Error; err != nil {
		JSON(c, http.StatusNotFound, 0, "settings not found")
		return
	}

	// Update the settings fields. You can use Model(&settings).Updates(newSettings) if preferred.
	settings = newSettings
	// Preserve the record's primary key if necessary, for example:
	// newSettings.ID = settings.ID

	if err := DB.Db.Save(&settings).Error; err != nil {
		JSON(c, http.StatusInternalServerError, 0, err.Error())
		return
	}
	JSON(c, http.StatusOK, 1, settings)
}

// @Summary Get Upload
// @Schemes
// @Description Returns upload endpoint information.
// @Tags upload
// @Accept json
// @Produce json
// @Success 200 {string} string "Upload endpoint OK"
// @Router /upload [get]
func getUpload(c *gin.Context) {
	if err := flow.ContinueUpload(c.Request); err != nil {
		JSON(c, http.StatusBadRequest, 0, err)
		log.Printf("Error: %s", err)
		return
	}
	JSON(c, http.StatusOK, 0, "Upload endpoint OK")
}

// @Summary Post Upload
// @Schemes
// @Description Handles file upload and saves the file.
// @Tags upload
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "File to upload"
// @Success 200 {string} string "File uploaded successfully"
// @Router /upload [post]
func postUpload(c *gin.Context) {
	if err := flow.StreamingReader(c.Request); err != nil {
		JSON(c, http.StatusBadRequest, 0, err)
		return
	}

	JSON(c, http.StatusOK, 0, "File uploaded successfully")
}
