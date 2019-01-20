package ahandlers

import (
	"github.com/Sirupsen/logrus"
	"github.com/dropbox/godropbox/container/set"
	"github.com/gin-gonic/gin"
	"github.com/pritunl/pritunl-cloud/data"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/demo"
	"github.com/pritunl/pritunl-cloud/event"
	"github.com/pritunl/pritunl-cloud/storage"
	"github.com/pritunl/pritunl-cloud/utils"
	"gopkg.in/mgo.v2/bson"
)

type storageData struct {
	Id        bson.ObjectId `json:"id"`
	Name      string        `json:"name"`
	Type      string        `json:"type"`
	Endpoint  string        `json:"endpoint"`
	Bucket    string        `json:"bucket"`
	AccessKey string        `json:"access_key"`
	SecretKey string        `json:"secret_key"`
	Insecure  bool          `json:"insecure"`
}

func storagePut(c *gin.Context) {
	if demo.Blocked(c) {
		return
	}

	db := c.MustGet("db").(*database.Database)
	dta := &storageData{}

	storeId, ok := utils.ParseObjectId(c.Param("store_id"))
	if !ok {
		utils.AbortWithStatus(c, 400)
		return
	}

	err := c.Bind(dta)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	store, err := storage.Get(db, storeId)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	store.Name = dta.Name
	store.Type = dta.Type
	store.Endpoint = dta.Endpoint
	store.Bucket = dta.Bucket
	store.AccessKey = dta.AccessKey
	store.SecretKey = dta.SecretKey
	store.Insecure = dta.Insecure

	fields := set.NewSet(
		"name",
		"type",
		"endpoint",
		"bucket",
		"access_key",
		"secret_key",
		"insecure",
	)

	errData, err := store.Validate(db)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	if errData != nil {
		c.JSON(400, errData)
		return
	}

	err = store.CommitFields(db, fields)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	go func() {
		db := database.GetDatabase()
		defer db.Close()

		err = data.Sync(db, store)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"error": err,
			}).Error("storage: Failed to sync storage")
		}
	}()

	event.PublishDispatch(db, "storage.change")

	c.JSON(200, store)
}

func storagePost(c *gin.Context) {
	if demo.Blocked(c) {
		return
	}

	db := c.MustGet("db").(*database.Database)
	dta := &storageData{
		Name: "New Storage",
	}

	err := c.Bind(dta)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	store := &storage.Storage{
		Name:      dta.Name,
		Type:      dta.Type,
		Endpoint:  dta.Endpoint,
		Bucket:    dta.Bucket,
		AccessKey: dta.AccessKey,
		SecretKey: dta.SecretKey,
		Insecure:  dta.Insecure,
	}

	errData, err := store.Validate(db)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	if errData != nil {
		c.JSON(400, errData)
		return
	}

	err = store.Insert(db)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	event.PublishDispatch(db, "storage.change")

	c.JSON(200, store)
}

func storageDelete(c *gin.Context) {
	if demo.Blocked(c) {
		return
	}

	db := c.MustGet("db").(*database.Database)

	storeId, ok := utils.ParseObjectId(c.Param("store_id"))
	if !ok {
		utils.AbortWithStatus(c, 400)
		return
	}

	err := storage.Remove(db, storeId)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	event.PublishDispatch(db, "storage.change")

	c.JSON(200, nil)
}

func storageGet(c *gin.Context) {
	db := c.MustGet("db").(*database.Database)

	storeId, ok := utils.ParseObjectId(c.Param("store_id"))
	if !ok {
		utils.AbortWithStatus(c, 400)
		return
	}

	store, err := storage.Get(db, storeId)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	if demo.IsDemo() {
		if store.AccessKey != "" {
			store.AccessKey = "demo"
		}
		if store.SecretKey != "" {
			store.SecretKey = "demo"
		}
	}

	c.JSON(200, store)
}

func storagesGet(c *gin.Context) {
	db := c.MustGet("db").(*database.Database)

	stores, err := storage.GetAll(db)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	if demo.IsDemo() {
		for _, store := range stores {
			if store.AccessKey != "" {
				store.AccessKey = "demo"
			}
			if store.SecretKey != "" {
				store.SecretKey = "demo"
			}
		}
	}

	c.JSON(200, stores)
}
