package mhandlers

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/dropbox/godropbox/container/set"
	"github.com/dropbox/godropbox/errors"
	"github.com/gin-gonic/gin"
	"github.com/pritunl/mongo-go-driver/bson"
	"github.com/pritunl/mongo-go-driver/bson/primitive"
	"github.com/pritunl/pritunl-zero/database"
	"github.com/pritunl/pritunl-zero/demo"
	"github.com/pritunl/pritunl-zero/endpoint"
	"github.com/pritunl/pritunl-zero/errortypes"
	"github.com/pritunl/pritunl-zero/event"
	"github.com/pritunl/pritunl-zero/utils"
)

type endpointData struct {
	Id    primitive.ObjectID `json:"id"`
	Name  string             `json:"name"`
	Roles []string           `json:"roles"`
}

type endpointsData struct {
	Endpoints []*endpoint.Endpoint `json:"endpoints"`
	Count     int64                `json:"count"`
}

func endpointPut(c *gin.Context) {
	if demo.Blocked(c) {
		return
	}

	db := c.MustGet("db").(*database.Database)
	data := &endpointData{}

	endpointId, ok := utils.ParseObjectId(c.Param("endpoint_id"))
	if !ok {
		utils.AbortWithStatus(c, 400)
		return
	}

	err := c.Bind(data)
	if err != nil {
		err = &errortypes.ParseError{
			errors.Wrap(err, "handler: Bind error"),
		}
		utils.AbortWithError(c, 500, err)
		return
	}

	endpt, err := endpoint.Get(db, endpointId)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	endpt.Name = data.Name
	endpt.Roles = data.Roles

	fields := set.NewSet(
		"name",
		"roles",
	)

	errData, err := endpt.Validate(db)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	if errData != nil {
		c.JSON(400, errData)
		return
	}

	err = endpt.CommitFields(db, fields)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	event.PublishDispatch(db, "endpoint.change")

	c.JSON(200, endpt)
}

func endpointPost(c *gin.Context) {
	if demo.Blocked(c) {
		return
	}

	db := c.MustGet("db").(*database.Database)
	data := &endpointData{
		Name: "New Endpoint",
	}

	err := c.Bind(data)
	if err != nil {
		err = &errortypes.ParseError{
			errors.Wrap(err, "handler: Bind error"),
		}
		utils.AbortWithError(c, 500, err)
		return
	}

	endpt := &endpoint.Endpoint{
		Name:  data.Name,
		Roles: data.Roles,
	}

	errData, err := endpt.Validate(db)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	if errData != nil {
		c.JSON(400, errData)
		return
	}

	err = endpt.Insert(db)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	event.PublishDispatch(db, "endpoint.change")

	c.JSON(200, endpt)
}

func endpointDelete(c *gin.Context) {
	if demo.Blocked(c) {
		return
	}

	db := c.MustGet("db").(*database.Database)

	endpointId, ok := utils.ParseObjectId(c.Param("endpoint_id"))
	if !ok {
		utils.AbortWithStatus(c, 400)
		return
	}

	err := endpoint.Remove(db, endpointId)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	event.PublishDispatch(db, "endpoint.change")

	c.JSON(200, nil)
}

func endpointsDelete(c *gin.Context) {
	if demo.Blocked(c) {
		return
	}

	db := c.MustGet("db").(*database.Database)
	dta := []primitive.ObjectID{}

	err := c.Bind(&dta)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	err = endpoint.RemoveMulti(db, dta)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	event.PublishDispatch(db, "endpoint.change")

	c.JSON(200, nil)
}

func endpointsGet(c *gin.Context) {
	db := c.MustGet("db").(*database.Database)
	page, _ := strconv.ParseInt(c.Query("page"), 10, 0)
	pageCount, _ := strconv.ParseInt(c.Query("page_count"), 10, 0)

	query := bson.M{}

	endpointId, ok := utils.ParseObjectId(c.Query("id"))
	if ok {
		query["_id"] = endpointId
	}

	name := strings.TrimSpace(c.Query("name"))
	if name != "" {
		query["$or"] = []*bson.M{
			&bson.M{
				"name": &bson.M{
					"$regex":   fmt.Sprintf(".*%s.*", name),
					"$options": "i",
				},
			},
			&bson.M{
				"key": &bson.M{
					"$regex":   fmt.Sprintf(".*%s.*", name),
					"$options": "i",
				},
			},
		}
	}

	typ := strings.TrimSpace(c.Query("type"))
	if typ != "" {
		query["type"] = typ
	}

	organization, ok := utils.ParseObjectId(c.Query("organization"))
	if ok {
		query["organization"] = organization
	}

	endpoints, count, err := endpoint.GetAllPaged(
		db, &query, page, pageCount)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	dta := &endpointsData{
		Endpoints: endpoints,
		Count:     count,
	}

	c.JSON(200, dta)
}
