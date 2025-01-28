package main

import (
	"database/sql"
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/logger"
	"github.com/gin-gonic/gin"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

func initDb() *sql.DB {
	db, err := sql.Open("sqlite3", "pushups.sqlite")
	if err != nil {
		log.Fatal(err)
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS pushups (
		    id INTEGER PRIMARY KEY AUTOINCREMENT,
		    count INTEGER NOT NULL,
		    time DATE NOT NULL
		);
	`)
	if err != nil {
		log.Fatal(err)
	}

	return db
}

func main() {
	db := initDb()

	r := gin.Default()

	corsRule := cors.Default()
	port := ":80"
	if os.Args[1] != "prod" {
		port = ":8080"
		corsRule = cors.New(cors.Config{AllowOrigins: []string{"https://pushups.cgreggescalante.com"}})
	}

	r.Use(corsRule)
	r.Use(logger.SetLogger())

	r.StaticFile("/", "index.html")

	r.POST("/log", func(c *gin.Context) {
		count := c.Query("reps")
		countInt, err := strconv.Atoi(count)
		if err != nil {
			c.String(http.StatusBadRequest, "Invalid count")
			return
		}

		t := time.Now().Unix()

		_, err = db.Exec("INSERT INTO pushups (count, time) VALUES (?, ?)", countInt, t)
		if err != nil {
			c.String(http.StatusInternalServerError, "Internal server error")
			return
		}
		c.JSON(http.StatusOK, []interface{}{countInt, t})
	})

	r.GET("/all", func(c *gin.Context) {
		rows, err := db.Query("SELECT count, time FROM pushups")
		if err != nil {
			log.Println("Error selecting pushups", err)
			c.String(http.StatusInternalServerError, "Internal server error")
			return
		}
		defer rows.Close()

		var items [][]interface{}

		var count int
		var logTime time.Time
		for rows.Next() {
			if err = rows.Scan(&count, &logTime); err != nil {
				log.Println("Error scanning pushups", err)
				c.String(http.StatusInternalServerError, "Internal server error")
				return
			}
			items = append(items, []interface{}{count, logTime.Unix()})
		}

		c.JSON(http.StatusOK, items)
	})

	log.Fatal(r.Run(port))
}
