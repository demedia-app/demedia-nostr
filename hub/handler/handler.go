package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sithumonline/demedia-nostr/relayer/storage/postgresql"
)

func Start(port string, m map[string]postgresql.PeerInfo) {
	r := gin.Default()
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Add("Access-Control-Allow-Origin", "*")
		c.Next()
	})

	v1 := r.Group("/v1")
	{
		v1.GET("/data", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"data": m})
		})
	}

	r.Run(port)
}
