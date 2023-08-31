package handler

import (
	"crypto/ecdsa"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/nbd-wtf/go-nostr"
	"github.com/sithumonline/demedia-nostr/relayer"
	"github.com/sithumonline/demedia-nostr/relayer/hashutil"
)

func Start(port string, relay relayer.Relay, pub *ecdsa.PublicKey) {
	r := gin.Default()
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Add("Access-Control-Allow-Origin", "*")
		c.Next()
	})

	v1 := r.Group("/v1")
	{
		v1.GET("/data", func(c *gin.Context) {
			events, err := relay.Storage().QueryEvents(&nostr.Filter{})
			for _, event := range events {
				if event.Kind == 1 && len(event.Tags) > 0 {
					tag := len(event.Tags) - 1
					if len(event.Tags[tag]) == 0 {
						continue
					}
					if tag != -1 && event.Tags[tag][0] == "hash" {
						b, err := hashutil.GetVerification(event.Tags[tag][1], hashutil.GetSha256([]byte(event.Content)), pub)
						if err != nil {
							log.Printf("failed to verify hash: %v", err)
						} else {
							event.Tags[tag][2] = strconv.FormatBool(b)
							bs := hashutil.GetSha256([]byte(hashutil.StringifyEvent(&event)))
							event.ID = fmt.Sprintf("%x", bs)
						}
					}
				}
			}
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
