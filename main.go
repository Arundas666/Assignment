package main

import (
	"assignment/handlers"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
)

func main() {
	router := gin.Default()
	datastore := handlers.NewDatastore()

	router.POST("/process", func(c *gin.Context) {
		handlers.Process(c, datastore)
	})

	go func() {
		for {
			time.Sleep(10 * time.Second)
			datastore.Mutex.Lock()
			now := time.Now()

			for key, expirationTime := range datastore.Expiry {
				if now.After(expirationTime) {
					delete(datastore.Data, key)
					delete(datastore.Expiry, key)
					fmt.Printf("Key %s expired and has been deleted from the datastore.\n", key)
				}
			}

			datastore.Mutex.Unlock()
		}
	}()
	router.Run(":8080")
	defer datastore.StopExpirationTimers()

}
