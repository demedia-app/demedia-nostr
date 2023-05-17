package handler

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/nbd-wtf/go-nostr"
	"github.com/sithumonline/demedia-nostr/relayer"
)

func Start(port string, relay relayer.Relay) {
	r := gin.Default()
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Add("Access-Control-Allow-Origin", "*")
		c.Next()
	})

	v1 := r.Group("/v1")
	{
		v1.GET("/data", func(c *gin.Context) {
			events, err := relay.Storage().QueryEvents(&nostr.Filter{})
			if err != nil {
				log.Printf("failed to get data from db: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, gin.H{"data": events})
		})
	}

	r.Run(port)
}
