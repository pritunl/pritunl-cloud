package ahandlers

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/pritunl/pritunl-cloud/auth"
	"github.com/pritunl/pritunl-cloud/authorizer"
	"github.com/pritunl/pritunl-cloud/config"
	"github.com/pritunl/pritunl-cloud/constants"
	"github.com/pritunl/pritunl-cloud/middlewear"
	"github.com/pritunl/pritunl-cloud/static"
	"github.com/pritunl/pritunl-cloud/utils"
)

func staticPath(c *gin.Context, pth string, cache bool) {
	pth = config.StaticRoot + pth

	file, ok := store.Files[pth]
	if !ok {
		utils.AbortWithStatus(c, 404)
		return
	}

	if constants.StaticCache && cache {
		c.Writer.Header().Add("Cache-Control", "public, max-age=86400")
		c.Writer.Header().Add("ETag", file.Hash)
	} else {
		c.Writer.Header().Add("Cache-Control",
			"no-cache, no-store, must-revalidate")
		c.Writer.Header().Add("Pragma", "no-cache")
		c.Writer.Header().Add("Expires", "0")
	}

	if strings.Contains(c.Request.Header.Get("Accept-Encoding"), "gzip") {
		c.Writer.Header().Add("Content-Encoding", "gzip")
		c.Data(200, file.Type, file.GzipData)
	} else {
		c.Data(200, file.Type, file.Data)
	}
}

func staticIndexGet(c *gin.Context) {
	authr := c.MustGet("authorizer").(*authorizer.Authorizer)
	if !authr.IsValid() {
		fastPth := auth.GetFastAdminPath()
		if fastPth != "" {
			c.Redirect(302, fastPth)
			return
		}

		c.Redirect(302, "/login")
		return
	}

	staticPath(c, "/index.html", false)
}

func staticLoginGet(c *gin.Context) {
	staticPath(c, "/login.html", false)
}

func staticLogoGet(c *gin.Context) {
	staticPath(c, "/logo.png", true)
}

func staticGet(c *gin.Context) {
	staticPath(c, "/static"+c.Params.ByName("path"), true)
}

func staticTestingGet(c *gin.Context) {
	pth := c.Params.ByName("path")
	if pth == "" {
		if c.Request.URL.Path == "/config.js" {
			pth = "config.js"
		} else if c.Request.URL.Path == "/logo.png" {
			pth = "logo.png"
		} else if c.Request.URL.Path == "/build.js" {
			pth = "build.js"
		} else if c.Request.URL.Path == "/login" {
			fastPth := auth.GetFastAdminPath()
			if fastPth != "" {
				c.Redirect(302, fastPth)
				return
			}

			c.Request.URL.Path = "/login.html"
			pth = "login.html"
		} else {
			authr := c.MustGet("authorizer").(*authorizer.Authorizer)
			if !authr.IsValid() {
				fastPth := auth.GetFastAdminPath()
				if fastPth != "" {
					c.Redirect(302, fastPth)
					return
				}

				c.Redirect(302, "/login")
				return
			}

			pth = "index.html"
		}
	}

	if strings.HasPrefix(c.Request.URL.Path, "/node_modules/") ||
		strings.HasPrefix(c.Request.URL.Path, "/jspm_packages/") {

		c.Writer.Header().Add("Cache-Control", "public, max-age=86400")
	} else {
		c.Writer.Header().Add("Cache-Control",
			"no-cache, no-store, must-revalidate")
		c.Writer.Header().Add("Pragma", "no-cache")
		c.Writer.Header().Add("Expires", "0")
	}

	c.Writer.Header().Add("Content-Type", static.GetMimeType(pth))

	gzipWriter := middlewear.NewGzipWriter(c)
	defer gzipWriter.Close()
	fileServer.ServeHTTP(gzipWriter, c.Request)
}
