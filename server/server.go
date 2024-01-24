package server

import (
	"database/sql"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
	"sync"
	"time"
)

type ApiServer struct {
	mu        sync.Mutex
	port      int64
	dbPath    string
	startTime int64
}

func StartApiServer(dbPath string, port int64) *ApiServer {
	apiServer := &ApiServer{
		port:      port,
		dbPath:    dbPath,
		startTime: time.Now().Unix(),
	}

	gin.SetMode(gin.ReleaseMode)
	r := gin.New()

	r.GET("/get_item/:from_timestamp/:to_timestamp", apiServer.GetItemHandler)
	r.GET("/get_latest_key", apiServer.GetLatestKey)

	if err := r.Run(fmt.Sprintf(":%d", port)); err != nil {
		panic(err)
	}

	return apiServer
}

func (apiServer *ApiServer) GetItemHandler(c *gin.Context) {
	fromTimestampUnixStr := c.Param("from_timestamp")
	toTimestampUnixStr := c.Param("to_timestamp")

	fmt.Printf("Received query: from %v to %v\n", fromTimestampUnixStr, toTimestampUnixStr)

	fromTimestampUnix, err := strconv.ParseInt(fromTimestampUnixStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	toTimestampUnix, err := strconv.ParseInt(toTimestampUnixStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	// return empty set if client requests data before server started
	if apiServer.startTime >= fromTimestampUnix {
		c.JSON(http.StatusOK, [][]byte{})
		return
	}

	// return empty set if client requests data which is in the future
	if time.Now().Unix() <= toTimestampUnix {
		c.JSON(http.StatusOK, [][]byte{})
		return
	}

	// Set up DB
	db, err := sql.Open("sqlite3", apiServer.dbPath)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	query := "SELECT data FROM messages WHERE strftime('%s', created) BETWEEN ? AND ?"
	apiServer.mu.Lock()
	rows, err := db.Query(query, fromTimestampUnixStr, toTimestampUnixStr)
	apiServer.mu.Unlock()
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	var messages [][]byte

	for rows.Next() {
		var data []byte
		err = rows.Scan(&data)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}

		messages = append(messages, data)
	}

	if len(messages) == 0 {
		c.JSON(http.StatusOK, [][]byte{})
	} else {
		c.JSON(http.StatusOK, messages)
	}
}

func (apiServer *ApiServer) GetLatestKey(c *gin.Context) {
	// Set up DB
	db, err := sql.Open("sqlite3", apiServer.dbPath)
	if err != nil {
		panic(err)
	}

	query := "SELECT MAX(created) FROM messages"

	var latestKey string
	rows, err := db.Query(query)
	defer rows.Close()

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	for rows.Next() {
		err := rows.Scan(&latestKey)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}
	}

	layout := "2006-01-02 15:04:05.99999-07:00"
	timestamp, err := time.Parse(layout, latestKey)
	if err != nil {
		panic(err)
	}

	// Unix-Timestamp extrahieren
	unixTimestamp := timestamp.Unix()

	c.JSON(http.StatusOK, gin.H{
		"latest_key": unixTimestamp,
	})
}
