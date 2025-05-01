package ahandlers

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/dropbox/godropbox/container/set"
	"github.com/dropbox/godropbox/errors"
	"github.com/gin-gonic/gin"
	"github.com/pritunl/mongo-go-driver/bson"
	"github.com/pritunl/mongo-go-driver/bson/primitive"
	"github.com/pritunl/pritunl-cloud/data"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/demo"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/event"
	"github.com/pritunl/pritunl-cloud/storage"
	"github.com/pritunl/pritunl-cloud/utils"
	"github.com/sirupsen/logrus"
)

type storageData struct {
	Id        primitive.ObjectID `json:"id"`
	Name      string             `json:"name"`
	Comment   string             `json:"comment"`
	Type      string             `json:"type"`
	Endpoint  string             `json:"endpoint"`
	Bucket    string             `json:"bucket"`
	AccessKey string             `json:"access_key"`
	SecretKey string             `json:"secret_key"`
	Insecure  bool               `json:"insecure"`
}

type storagesData struct {
	Storages []*storage.Storage `json:"storages"`
	Count    int64              `json:"count"`
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
	store.Comment = dta.Comment
	store.Type = dta.Type
	store.Endpoint = dta.Endpoint
	store.Bucket = dta.Bucket
	store.AccessKey = dta.AccessKey
	store.SecretKey = dta.SecretKey
	store.Insecure = dta.Insecure

	fields := set.NewSet(
		"name",
		"comment",
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

		event.PublishDispatch(db, "image.change")
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
		Comment:   dta.Comment,
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

	go func() {
		db := database.GetDatabase()
		defer db.Close()

		err = data.Sync(db, store)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"error": err,
			}).Error("storage: Failed to sync storage")
		}

		event.PublishDispatch(db, "image.change")
	}()

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

func storagesDelete(c *gin.Context) {
	if demo.Blocked(c) {
		return
	}

	db := c.MustGet("db").(*database.Database)
	data := []primitive.ObjectID{}

	err := c.Bind(&data)
	if err != nil {
		err = &errortypes.ParseError{
			errors.Wrap(err, "handler: Bind error"),
		}
		utils.AbortWithError(c, 500, err)
		return
	}

	err = storage.RemoveMulti(db, data)
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

	page, _ := strconv.ParseInt(c.Query("page"), 10, 0)
	pageCount, _ := strconv.ParseInt(c.Query("page_count"), 10, 0)

	query := bson.M{}

	storageId, ok := utils.ParseObjectId(c.Query("id"))
	if ok {
		query["_id"] = storageId
	}

	name := strings.TrimSpace(c.Query("name"))
	if name != "" {
		query["name"] = &bson.M{
			"$regex":   fmt.Sprintf(".*%s.*", regexp.QuoteMeta(name)),
			"$options": "i",
		}
	}

	organization, ok := utils.ParseObjectId(c.Query("organization"))
	if ok {
		query["organization"] = organization
	}

	datacenter, ok := utils.ParseObjectId(c.Query("datacenter"))
	if ok {
		query["datacenter"] = datacenter
	}

	comment := strings.TrimSpace(c.Query("comment"))
	if comment != "" {
		query["comment"] = &bson.M{
			"$regex":   fmt.Sprintf(".*%s.*", comment),
			"$options": "i",
		}
	}

	stores, count, err := storage.GetAllPaged(db, &query, page, pageCount)
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

	data := &storagesData{
		Storages: stores,
		Count:    count,
	}

	c.JSON(200, data)
}
