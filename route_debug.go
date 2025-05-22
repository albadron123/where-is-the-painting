package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

// IMPORTANT: this root is dedicated to debug purpose!
// It's used to test different kinds of behaviours!!!
func get_DEBUG_PAGE(c *gin.Context) {
	c.HTML(http.StatusOK, "dropdown_test.html", gin.H{})
}

func post_DEBUG(c *gin.Context) {
	//obj := ProfileForm{}

	/*
		if strings.Contains(c.GetHeader("content-type"), "multipart") {
			c.JSON(401, gin.H{"test": "test"})
			return
		}
	*/

	/*
		// Multipart form
		if err := c.ShouldBind(&obj); err != nil {
			c.JSON(406, gin.H{"error": err.Error()})
			return
		}
	*/
	in := []byte(c.Request.FormValue("json"))
	file, _ := c.FormFile("file")

	type Res struct {
		Hello string `json:"hello"`
	}
	var res Res
	err := json.Unmarshal(in, &res)

	if err != nil {
		fmt.Println(err.Error())
		return
	} else {
		fmt.Println(res.Hello)
	}

	/*
		err := c.BindJSON(obj)
		if err != nil {
			fmt.Println(err.Error())
		}
	*/
	//file := form.File["file"]
	fmt.Println(file.Filename)
	err = c.SaveUploadedFile(file, "./assets/"+file.Filename)
	if err != nil {
		fmt.Println(err.Error())
		fmt.Println("error")
	}
	/*
		for _, file := range files {
			log.Println(file.Filename)

			// Upload the file to specific dst.
		}
	*/
	//post_DEBUG_RAW(c)
}

// prints the size of the raw post data in console
func post_DEBUG_RAW(c *gin.Context) {
	raw, err := c.GetRawData()
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	fmt.Println("size of message: ", len(raw), " bytes")
}
