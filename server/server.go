package server

import (
	"database/sql"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
)

type ApiServer struct {
	port   int64
	dbPath string
}

func StartApiServer(dbPath string, port int64) *ApiServer {
	apiServer := &ApiServer{
		port:   port,
		dbPath: dbPath,
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

	println(fmt.Sprintf("Received query: from %v to %v", fromTimestampUnixStr, toTimestampUnixStr))

	// Set up DB
	db, err := sql.Open("sqlite3", apiServer.dbPath)
	if err != nil {
		panic(err)
	}

	query := "SELECT data FROM messages WHERE strftime('%s', created) BETWEEN ? AND ?"

	rows, err := db.Query(query, fromTimestampUnixStr, toTimestampUnixStr)
	if err != nil {
		panic(err)
	}
	var data []byte

	var dataList []byte
	for rows.Next() {
		err = rows.Scan(&data)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}
		dataList = append(dataList, data...)
	}
	rows.Close()
	c.Data(http.StatusOK, "application/json", dataList)
}

func (apiServer *ApiServer) GetLatestKey(c *gin.Context) {
	// Set up DB
	db, err := sql.Open("sqlite3", apiServer.dbPath)
	if err != nil {
		panic(err)
	}

	query := "SELECT MAX(created) FROM messages"

	var latestKey string
	err = db.QueryRow(query).Scan(&latestKey)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
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
