package ahandlers

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/dropbox/godropbox/container/set"
	"github.com/gin-gonic/gin"
	"github.com/pritunl/mongo-go-driver/bson"
	"github.com/pritunl/mongo-go-driver/bson/primitive"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/demo"
	"github.com/pritunl/pritunl-cloud/event"
	"github.com/pritunl/pritunl-cloud/user"
	"github.com/pritunl/pritunl-cloud/utils"
)

type userData struct {
	Id             primitive.ObjectID `json:"id"`
	Type           string             `json:"type"`
	Username       string             `json:"username"`
	Password       string             `json:"password"`
	Comment        string             `json:"comment"`
	Roles          []string           `json:"roles"`
	Administrator  string             `json:"administrator"`
	Permissions    []string           `json:"permissions"`
	GenerateSecret bool               `json:"generate_secret"`
	Disabled       bool               `json:"disabled"`
	ActiveUntil    time.Time          `json:"active_until"`
}

type usersData struct {
	Users []*user.User `json:"users"`
	Count int64        `json:"count"`
}

func userGet(c *gin.Context) {
	if demo.IsDemo() {
		usr := demo.Users[0]
		usr.LastActive = time.Now()
		c.JSON(200, usr)
		return
	}

	db := c.MustGet("db").(*database.Database)

	userId, ok := utils.ParseObjectId(c.Param("user_id"))
	if !ok {
		utils.AbortWithStatus(c, 400)
		return
	}

	usr, err := user.Get(db, userId)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	usr.Secret = ""

	c.JSON(200, usr)
}

func userPut(c *gin.Context) {
	if demo.Blocked(c) {
		return
	}

	db := c.MustGet("db").(*database.Database)
	data := &userData{}

	userId, ok := utils.ParseObjectId(c.Param("user_id"))
	if !ok {
		utils.AbortWithStatus(c, 400)
		return
	}

	err := c.Bind(data)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	usr, err := user.Get(db, userId)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	showSecret := false
	if usr.Type != data.Type {
		if data.Type == user.Api {
			usr.GenerateToken()
			showSecret = true
		} else {
			usr.Token = ""
			usr.Secret = ""
		}
	}

	usr.Type = data.Type
	usr.Username = data.Username
	usr.Comment = data.Comment
	usr.Roles = data.Roles
	usr.Administrator = data.Administrator
	usr.Permissions = data.Permissions
	usr.Disabled = data.Disabled
	usr.ActiveUntil = data.ActiveUntil

	if usr.Disabled {
		usr.ActiveUntil = time.Time{}
	}

	if usr.Type == user.Api && data.GenerateSecret {
		usr.GenerateToken()
		showSecret = true
	}

	fields := set.NewSet(
		"type",
		"token",
		"secret",
		"username",
		"comment",
		"roles",
		"administrator",
		"permissions",
		"disabled",
		"active_until",
	)

	if usr.Type == user.Local && data.Password != "" {
		err = usr.SetPassword(data.Password)
		if err != nil {
			utils.AbortWithError(c, 500, err)
			return
		}

		fields.Add("password")
	} else if usr.Type != user.Local && usr.Password != "" {
		usr.Password = ""
		fields.Add("password")
	}

	errData, err := usr.Validate(db)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	if errData != nil {
		c.JSON(400, errData)
		return
	}

	errData, err = usr.SuperExists(db)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	if errData != nil {
		c.JSON(400, errData)
		return
	}

	err = usr.CommitFields(db, fields)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	event.PublishDispatch(db, "user.change")

	if !showSecret {
		usr.Secret = ""
	}

	c.JSON(200, usr)
}

func userPost(c *gin.Context) {
	if demo.Blocked(c) {
		return
	}

	db := c.MustGet("db").(*database.Database)
	data := &userData{}

	err := c.Bind(data)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	usr := &user.User{
		Type:          data.Type,
		Username:      data.Username,
		Comment:       data.Comment,
		Roles:         data.Roles,
		Administrator: data.Administrator,
		Permissions:   data.Permissions,
		Disabled:      data.Disabled,
		ActiveUntil:   data.ActiveUntil,
	}

	if usr.Disabled {
		usr.ActiveUntil = time.Time{}
	}

	if usr.Type == user.Local && data.Password != "" {
		err = usr.SetPassword(data.Password)
		if err != nil {
			utils.AbortWithError(c, 500, err)
			return
		}
	}

	if usr.Type == user.Api {
		usr.GenerateToken()
	}

	errData, err := usr.Validate(db)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	if errData != nil {
		c.JSON(400, errData)
		return
	}

	err = usr.Insert(db)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	event.PublishDispatch(db, "user.change")

	c.JSON(200, usr)
}

func usersGet(c *gin.Context) {
	if demo.IsDemo() {
		for _, usr := range demo.Users {
			usr.LastActive = time.Now()
		}

		data := &usersData{
			Users: demo.Users,
			Count: int64(len(demo.Users)),
		}

		c.JSON(200, data)
		return
	}

	db := c.MustGet("db").(*database.Database)

	page, _ := strconv.ParseInt(c.Query("page"), 10, 0)
	pageCount, _ := strconv.ParseInt(c.Query("page_count"), 10, 0)

	query := bson.M{}

	userId, ok := utils.ParseObjectId(c.Query("id"))
	if ok {
		query["_id"] = userId
	}

	username := strings.TrimSpace(c.Query("username"))
	if username != "" {
		query["username"] = &bson.M{
			"$regex":   fmt.Sprintf(".*%s.*", regexp.QuoteMeta(username)),
			"$options": "i",
		}
	}

	role := strings.TrimSpace(c.Query("role"))
	if role != "" {
		query["roles"] = role
	}

	typ := strings.TrimSpace(c.Query("type"))
	if typ != "" {
		query["type"] = typ
	}

	administrator := c.Query("administrator")
	switch administrator {
	case "true":
		query["administrator"] = "super"
		break
	case "false":
		query["administrator"] = ""
		break
	}

	disabled := c.Query("disabled")
	switch disabled {
	case "true":
		query["disabled"] = true
		break
	case "false":
		query["disabled"] = false
		break
	}

	users, count, err := user.GetAll(db, &query, page, pageCount)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	for _, usr := range users {
		usr.Secret = ""
	}

	data := &usersData{
		Users: users,
		Count: count,
	}

	c.JSON(200, data)
}

func usersDelete(c *gin.Context) {
	if demo.Blocked(c) {
		return
	}

	db := c.MustGet("db").(*database.Database)
	data := []primitive.ObjectID{}

	err := c.Bind(&data)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	errData, err := user.Remove(db, data)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	if errData != nil {
		c.JSON(400, errData)
		return
	}

	event.PublishDispatch(db, "user.change")

	c.JSON(200, nil)
}
