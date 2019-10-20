package apilist

import (
	models "github.com/hiromaily/go-gin-wrapper/pkg/models/mysql"
	//gin "gopkg.in/gin-gonic/gin.v1"
	"github.com/gin-gonic/gin"
	//lg "github.com/hiromaily/golibs/log"
	"net/http"

	"github.com/hiromaily/go-gin-wrapper/pkg/libs/response/html"
)

// IndexAction is top page for API List (react version)
func IndexAction(c *gin.Context) {
	//return header and key

	//Get User ids
	type UserID struct {
		ID int
	}
	var ids []UserID

	err := models.GetDB().GetUserIds(&ids)
	if err != nil {
		c.AbortWithError(500, err)
		return
	}

	//View
	res := gin.H{
		"title":    "API List Page",
		"navi_key": "/apilist/",
		"ids":      ids,
	}
	c.HTML(http.StatusOK, "pages/apilist/index.tmpl", html.Response(res))
}

// Index2Action is top page for API List (this is old version)
func Index2Action(c *gin.Context) {
	//debug log
	//debugContext(c)

	//Get User ids
	type UserID struct {
		ID int
	}
	var ids []UserID

	err := models.GetDB().GetUserIds(&ids)
	if err != nil {
		c.AbortWithError(500, err)
		return
	}

	//View
	res := gin.H{
		"title":    "API List Page",
		"navi_key": "/apilist/",
		"ids":      ids,
	}
	c.HTML(http.StatusOK, "pages/apilist/index2.tmpl", html.Response(res))
}