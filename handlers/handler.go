package handlers

import (
	response "assignment/Response"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type Command struct {
	Value string `json:"command"`
}

type DataStore struct {
	Data    map[string]string
	Expiry  map[string]time.Time
	Mutex   sync.RWMutex
	StopCh  chan struct{}
	WaitGrp sync.WaitGroup
}

func NewDatastore() *DataStore {
	return &DataStore{
		Data:    make(map[string]string),
		Expiry:  make(map[string]time.Time),
		Mutex:   sync.RWMutex{},
		StopCh:  make(chan struct{}),
		WaitGrp: sync.WaitGroup{},
	}
}

func (ds *DataStore) StartExpirationTimer(key string, expiration int) {
	ds.WaitGrp.Add(1)
	go func() {
		defer ds.WaitGrp.Done()
		select {
		case <-time.After(time.Duration(expiration) * time.Second):
			ds.Mutex.Lock()
			delete(ds.Data, key)
			delete(ds.Expiry, key)
			ds.Mutex.Unlock()
			fmt.Printf("Key %s expired and has been deleted from the datastore.\n", key)
		case <-ds.StopCh:
			return
		}
	}()
}
func (ds *DataStore) StopExpirationTimers() {
	close(ds.StopCh)
	ds.WaitGrp.Wait()
}

func Process(c *gin.Context, datastore *DataStore) {
	var command Command
	if err := c.ShouldBindJSON(&command); err != nil {
		return
	}
	words := strings.Fields(command.Value)
	switch words[0] {
	case "SET":

		if len(words) < 3 {
			response.ErrorResponse(c, 400, "Invalid SET command. Expected format: SET <key> <value> <expiry time>? <condition>?", errors.New("invalid input"), nil)

			return
		}
		key := words[1]
		value := words[2]
		var expiration int
		var condition string

		for i := 3; i < len(words); i++ {
			switch words[i] {

			case "EX":
				if i+1 < len(words) {
					expirationStr := words[i+1]
					expiration, _ = strconv.Atoi(expirationStr)
					fmt.Println("INT IS", expiration)
					i = i + 1
				} else {
					response.ErrorResponse(c, 400, "Invalid SET command. Missing expiration time after EX.", errors.New("invalid input"), words[i])
					return
				}
			case "XX", "NX":
				condition = words[i]
				fmt.Println(condition, "CONDITOION")
			default:
				response.ErrorResponse(c, 400, "Invalid SET command. Unknown parameter:", errors.New("invalid input"), words[i])
				return
			}
		}
		datastore.Mutex.Lock()
		defer datastore.Mutex.Unlock()

		if condition == "XX" {
			// Only set the key if it already exists
			if _, exists := datastore.Data[key]; !exists {
				response.ErrorResponse(c, 500, "Key does not exist. Cannot perform XX operation.", errors.New("internal error"), nil)

				return
			}
		} else if condition == "NX" {
			// Only set the key if it does not already exist
			if _, exists := datastore.Data[key]; exists {
				response.ErrorResponse(c, 500, "Key already exists. Cannot perform NX operation.", errors.New("internal error"), nil)
				return
			}
		}

		if expiration > 0 {

			datastore.Expiry[key] = time.Now().Add(time.Duration(expiration) * time.Second)
			datastore.StartExpirationTimer(key, expiration)

		}
		datastore.Data[key] = value
		response.SuccessResponse(c, 200, "Data stored successfully", key)
		fmt.Printf("SET command processed. Key: %s, Value: %s\n", key, value)

	case "GET":
		if len(words) !=2 {
			response.ErrorResponse(c, 400, "Invalid GET command. Expected format: GET <key>", errors.New("invalid input"), nil)

			return
		}
		key := words[1]
		if _, exists := datastore.Data[key]; !exists {
			response.ErrorResponse(c, 500, "Key does not exist. Cannot perform XX operation.", errors.New("internal error"), nil)
			return
		}
		response.SuccessResponse(c, 200, "Data stored successfully", datastore.Data[key])



	default:
		response.ErrorResponse(c, 400, "Invalid  command. Unknown parameter:", errors.New("invalid input"), words[0])
		return

	}

}
