package main

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func getRegisterMuseum_PAGE(c *gin.Context) {
	fmt.Println("hello!")
	c.HTML(http.StatusOK, "register_museum.html", gin.H{})
}

func getMuseum_PAGE(c *gin.Context) {

	userIdFloat, exists := c.Get("user_id")
	if !exists {
		c.HTML(http.StatusUnauthorized, "unauthorized.html", gin.H{})
		return
	}
	userId := int(userIdFloat.(float64))

	museum_id, err := strconv.Atoi(c.Param("museum_id"))
	if err != nil {
		c.HTML(http.StatusBadRequest, "not_found.html", gin.H{})
		return
	}

	m := Museum{}
	err = db.Where("id = ?", museum_id).First(&m).Error
	if err != nil {
		c.HTML(http.StatusNotFound, "not_found.html", gin.H{})
		return
	}

	var right Right
	err = db.Where("user_id = ? and museum_id = ?", userId, museum_id).First(&right).Error
	if err != nil {
		c.HTML(http.StatusUnauthorized, "not_permitted", gin.H{})
		return
	}

	type UsersList struct {
		Id      int
		Login   string
		IsAdmin bool
	}
	usersList := []UsersList{}
	err = db.Raw("select users.id as id, users.login as login, is_admin  from ((select user_id, is_admin from rights where museum_id = ?) join users on user_id = users.id)", museum_id).Find(&usersList).Error
	if err != nil {
		fmt.Print("CRITICAL Error!!!!!!!!!!!!!!!!!!!!!!!!!!!")
	}
	c.HTML(http.StatusOK, "museum_page.html", gin.H{"museum_id": museum_id, "museum_name": m.Name, "users": usersList, "am_admin": right.IsAdmin, "my_id": userId})
}

func getRegisterAuthor_PAGE(c *gin.Context) {
	c.HTML(http.StatusOK, "register_author.html", gin.H{})
}

func getMain_PAGE(c *gin.Context) {
	c.HTML(http.StatusOK, "index.html", gin.H{})
}

func registerPainting_PAGE(c *gin.Context) {
	museumId, err := strconv.Atoi(c.Param("museum_id"))
	if err != nil {
		c.HTML(http.StatusBadRequest, "not_found.html", gin.H{})
		return
	}
	c.HTML(http.StatusOK, "register_painting.html", gin.H{"museum_id": museumId})
}

func success_PAGE(c *gin.Context) {
	c.HTML(http.StatusOK, "success.html", gin.H{})
}
