package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
)

func requireAuth(c *gin.Context) {

	tokenString, err := c.Cookie("Auth")

	if err != nil {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte("top-secret"), nil
	})

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		if float64(time.Now().Unix()) > claims["exp"].(float64) {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		fmt.Println(claims["sub"])
		c.Set("user_id", claims["sub"])
	} else {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	c.Next()
}

func requireAuthPAGE(c *gin.Context) {

	tokenString, err := c.Cookie("Auth")

	if err != nil {
		c.HTML(http.StatusUnauthorized, "unauthorized.html", gin.H{})
		c.Abort()
		return
	}

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte("top-secret"), nil
	})

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		if float64(time.Now().Unix()) > claims["exp"].(float64) {
			c.HTML(http.StatusUnauthorized, "unauthorized.html", gin.H{})
			c.Abort()
			return
		}
		fmt.Println(claims["sub"])
		c.Set("user_id", claims["sub"])
	} else {
		c.HTML(http.StatusUnauthorized, "unauthorized.html", gin.H{})
		c.Abort()
		return
	}
	c.Next()
}
