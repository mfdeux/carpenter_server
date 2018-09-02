package main

import (
	"bytes"
	"net/http"
	"strings"
	"time"

	"github.com/boltdb/bolt"
	"github.com/gin-gonic/gin"
	uuid "github.com/satori/go.uuid"
)

func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
	}
}

func requestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("X-Request-Id", uuid.NewV4().String())
	}
}

func tokenAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authToken := c.Query("token")
		if len(authToken) == 0 {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"status":  http.StatusForbidden,
				"message": "Permission denied",
			})
			return
		}
		if authToken != authTokenFlag {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"status":  http.StatusUnauthorized,
				"message": "Unauthorized",
			})
			return
		}
		return
	}
}

func usageStatsMiddleware(granularity int) gin.HandlerFunc {
	return func(c *gin.Context) {
		if strings.Contains(c.Request.URL.Path, "stats") || strings.Contains(c.Request.URL.Path, "health") {
			return
		}
		visitHit(granularity)
	}
}

func getBeginHour(offset int) time.Time {
	var now time.Time
	if offset == 0 {
		now = time.Now()
	} else {
		now = time.Now().Add(time.Duration(offset) * time.Hour)
	}
	return time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), 0, 0, 0, now.Location())
}

func visitHit(granularity int) {
	currentInterval := getLastInterval(time.Now(), granularity)
	db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("usage"))
		v := b.Get([]byte(currentInterval.Format(time.RFC3339)))
		if len(v) > 0 {
			intV := int(v[0])
			intV++
			v = []byte(string(intV))
		} else {
			v = []byte(string(1))
		}
		err := b.Put([]byte(currentInterval.Format(time.RFC3339)), v)
		return err
	})
}

func getLastInterval(date time.Time, interval int) time.Time {
	var modulo int
	modulo = date.Minute() % interval
	return time.Date(date.Year(), date.Month(), date.Day(), date.Hour(), date.Minute()-modulo, 0, 0, date.Location())
}

func getIntervalSums(granularity int, duration time.Duration) int {
	now := time.Now()
	startTime := now.Add(-duration)
	currentInterval := getLastInterval(now, granularity)
	startInterval := getLastInterval(startTime, granularity)
	count := 0
	db.View(func(tx *bolt.Tx) error {
		c := tx.Bucket([]byte("usage")).Cursor()
		min := []byte(startInterval.Format(time.RFC3339))
		max := []byte(currentInterval.Format(time.RFC3339))
		for k, v := c.Seek(min); k != nil && bytes.Compare(k, max) <= 0; k, v = c.Next() {
			if len(v) > 0 {
				count += int(v[0])
			}
		}
		return nil
	})
	return count
}
