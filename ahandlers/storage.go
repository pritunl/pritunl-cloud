package ahandlers

import (
	"github.com/dropbox/godropbox/container/set"
	"github.com/gin-gonic/gin"
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
	data := &storageData{}

	storeId, ok := utils.ParseObjectId(c.Param("store_id"))
	if !ok {
		utils.AbortWithStatus(c, 400)
		return
	}

	err := c.Bind(data)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	store, err := storage.Get(db, storeId)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	store.Name = data.Name
	store.Type = data.Type
	store.Endpoint = data.Endpoint
	store.Bucket = data.Bucket
	store.AccessKey = data.AccessKey
	store.SecretKey = data.SecretKey
	store.Insecure = data.Insecure

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

	event.PublishDispatch(db, "storage.change")

	c.JSON(200, store)
}

func storagePost(c *gin.Context) {
	if demo.Blocked(c) {
		return
	}

	db := c.MustGet("db").(*database.Database)
	data := &storageData{
		Name: "New Storage",
	}

	err := c.Bind(data)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	store := &storage.Storage{
		Name:      data.Name,
		Type:      data.Type,
		Endpoint:  data.Endpoint,
		Bucket:    data.Bucket,
		AccessKey: data.AccessKey,
		SecretKey: data.SecretKey,
		Insecure:  data.Insecure,
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

	c.JSON(200, store)
}

func storagesGet(c *gin.Context) {
	db := c.MustGet("db").(*database.Database)

	stores, err := storage.GetAll(db)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	c.JSON(200, stores)
}
