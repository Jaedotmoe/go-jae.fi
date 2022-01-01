package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"log"
	"html/template"
	"errors"

	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	"github.com/russross/blackfriday"
)

// For the future
type Website struct {
	Name  string `json:"name"`
	Url   string `json:"url"`
	Owner string `json:"owner"`
}

type Members struct {
	Members []Website `json:"members"`
}

type Post struct {
    Title   string
    Content template.HTML
}

func main() {
	// Web service
	r := gin.New()
	r.Use(gin.Recovery())

	// Streaming
	stream_online := false

	// Compression
	r.Use(gzip.Gzip(gzip.DefaultCompression))

	r.Delims("{{", "}}")

	r.GET("/me", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"name":       "Jae",
			"familyname": "Lo Presti",
			"job":        "Back-end Developer",
			"birthday":   "2001-04-28",
			"location":   "Helsinki, Finland, Europe, Earth, Alpha Quadrant",
			"email":      "me@jae.fi",
			"matrix":     "@me:jae.fi",
			"fediverse":  "@jae@mastodon.tedomum.net",
			"pronouns":   "She/Her",
			"stream_online": stream_online,
		})
	})

	r.GET("/webring/members", func(c *gin.Context) {
		jsonfile, err := ioutil.ReadFile("./webring/members.json")
		if err != nil {
			c.JSON(500, gin.H{
				"error": "Failed to read members",
			})
			return
		}
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
        c.Writer.Header().Set("Access-Control-Allow-Credentials", "false")
        c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
        c.Writer.Header().Set("Access-Control-Allow-Methods", "GET")

		c.Data(http.StatusOK, "application/json", jsonfile)
	})

	r.LoadHTMLGlob("templates/**/*.tmpl")
	r.Static("/assets", "./static")
	r.Static("/.well-known", "./wellknown")

	r.GET("/", func(c *gin.Context) {
		jsonfile, err := os.Open("./webring/members.json")
		if err != nil {
			fmt.Println(err)
		}
		var members Members

		byteValue, _ := ioutil.ReadAll(jsonfile)

		defer jsonfile.Close()

		json.Unmarshal(byteValue, &members)

		var previousSite, nextSite, randomsite string
		currentSite := "https://jae.fi"
		randSite := rand.Intn(len(members.Members) - 1)

		for i := 0; i < len(members.Members); i++ {
			if members.Members[i].Url == currentSite {
				if i-1 < 0 {
					previousSite = members.Members[len(members.Members)-1].Url
				} else {
					previousSite = members.Members[i-1].Url
				}

				if i+1 > len(members.Members) {
					nextSite = members.Members[i-1].Url
				} else {
					nextSite = members.Members[i+1].Url
				}

				randomsite = members.Members[randSite].Url
			}
		}

		c.HTML(http.StatusOK, "home/index.tmpl", gin.H{
			"title":        "Main page",
			"currentsite":  "Jae's Website",
			"currentowner": "Jae Lo Presti",
			"prevsite":     previousSite,
			"nextsite":     nextSite,
			"randomsite":   randomsite,
			"path":         c.FullPath(),
			"stream_online": stream_online,
		})
	})

	// Get routes
	r.GET("/:page", func(c *gin.Context) {
		requestedPage := c.Param("page")
		contentLocation := "content/" + requestedPage + ".md"

		// Check if file exists
		if _, err := os.Stat("content/" + requestedPage + ".md"); errors.Is(err, os.ErrNotExist){
			contentLocation = "content/404.md"
		}

		postContent, err := ioutil.ReadFile(contentLocation)
		if err != nil {
			log.Fatal(err)
		}
		
		postHTML := template.HTML(blackfriday.MarkdownCommon([]byte(postContent)))

		post := Post{Title: requestedPage, Content: postHTML}

		c.HTML(http.StatusOK, "globals/complete.tmpl", gin.H{
			"title": post.Title,
			"content": post.Content,
			"stream_online": stream_online,
		})

	})

	r.Run(":2021")
}
