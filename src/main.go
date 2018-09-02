package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/boltdb/bolt"
	"github.com/cnjack/throttle"
	"github.com/gin-gonic/gin"
)

func setupDB() (*bolt.DB, error) {
	db, err := bolt.Open("stats.db", 0666, nil)
	if err != nil {
		log.Fatal(err)
	}
	err = db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte("usage"))
		if err != nil {
			return fmt.Errorf("could not create root bucket: %v", err)
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("could not set up buckets, %v", err)
	}
	fmt.Println("DB Setup Done")
	return db, nil
}

var hostFlag string
var portFlag string
var throttleFlag int64
var behindProxy bool
var authTokenFlag string
var trackUsageFlag bool
var db *bolt.DB

func init() {
	flag.StringVar(&hostFlag, "host", "127.0.0.1", "Host the server should run on")
	flag.StringVar(&portFlag, "port", "9000", "Port the server should run on")
	flag.Int64Var(&throttleFlag, "throttle", 60, "Requests per minute allowed from IP")
	flag.BoolVar(&behindProxy, "proxy", false, "Whether the server is behind a proxy")
	flag.StringVar(&authTokenFlag, "token", "", "Token used to authorized requests")
	flag.BoolVar(&trackUsageFlag, "stats", true, "Whether to track usage statistics")
	flag.Parse()
}

func main() {
	router := gin.Default()
	router.Use(CORS())
	router.Use(requestIDMiddleware())
	if trackUsageFlag == true {
		var err error
		db, err = setupDB()
		if err != nil {
			log.Fatal(err)
		}
		router.Use(usageStatsMiddleware(5))
	}
	if authTokenFlag != "" {
		router.Use(tokenAuthMiddleware())
	}
	router.Use(throttle.Policy(&throttle.Quota{
		Limit:  uint64(throttleFlag),
		Within: time.Minute,
	}))
	router.POST("/", handleActionRequest)
	router.GET("/health", handleHealthCheck)
	if trackUsageFlag == true {
		router.GET("/stats", handleStatsRequest)
	}
	connString := fmt.Sprintf("%s:%s", hostFlag, portFlag)
	router.Run(connString)
}
