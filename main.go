package main

import (
	"context" //transactions
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

type painting struct {
	Id             int    `json:"id"`
	Title          string `json:"title"`
	CreationYear   string `json:"creation_year"`
	WhereToFind    string `json:"where_to_find"`
	PictureAddress string `json:"picture_address"`
}

type author struct {
	Id        int    `json:"id"`
	Name      string `json:"name"`
	BirthYear *int   `json:"birth_year"`
	DeathYear *int   `json:"death_year"`
	Biography string `json:"biography"`
}

type rights struct {
	UserId          int
	MuseumId        int
	GiveRights      bool
	ChangePaintings bool
}

var db *sql.DB

func main() {
	router := gin.Default()
	router.Static("/assets", "./assets")

	router.GET("/", getLanding)

	router.GET("/paintings_:request", searchPainting)
	router.GET("/authors_:request", searchAuthor)
	router.POST("/login", login)
	router.POST("/register", register)

	router.POST("/register_museum", requireAuth, postRegisterMuseum)

	router.POST("/museum:museum_id/create_painting", requireAuth, postPainting)
	router.GET("/museum:museum_id/all_paintings", getMuseumPaintings)

	//router.GET("/museum:museum_id/rights", requireAuth, getAllUsersRights)
	router.POST("/museum:museum_id/rights", requireAuth, postUserRights)
	router.PUT("/museum:museum_id/rights", requireAuth, changeUserRights)
	router.DELETE("/museum:museum_id/rights", requireAuth, deleteUserRights)

	//router.PUT("/painting:painting_id/change_painting", requireAuth, changePainting)
	//router.DELETE("/painting:painting_id/delete_painting", requireAuth, deletePainting)

	router.POST("/favorite", requireAuth, postFavorite)
	router.DELETE("/favorite", requireAuth, deleteFavorite)

	router.GET("/login_info", requireAuth, getLoginInfo)

	router.LoadHTMLGlob("templates/*.html")

	connStr := "user=postgres password=pass dbname=Paintings_Web_App sslmode=disable"
	var err error
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		panic(err)
	}
	fmt.Println("DB", db)
	defer db.Close()

	fmt.Println("Starting server...")
	router.Run("localhost:8080")
}

func getLanding(c *gin.Context) {
	c.HTML(http.StatusOK, "index.html", gin.H{})
}

func searchPainting(c *gin.Context) {
	type body struct {
		Id             int    `json:"id"`
		Title          string `json:"title"`
		CreationYear   string `json:"creation_year"`
		WhereToFind    string `json:"where_to_find"`
		PictureAddress string `json:"picture_address"`
		MuseumId       int    `json:"museum_id"`
		AuthorId       int    `json:"author_id"`
		AuthorName     string `json:"author_name"`
		MuseumName     string `json:"museum_name"`
	}
	request := c.Param("request")
	fmt.Println(request)
	sql := fmt.Sprintf("select p.*, a.name, m.name from ((select * from paintings where LOWER(title) like LOWER('%%%s%%')) as p join authors as a on p.author_id=a.id) join museums as m on p.museum_id = m.id", request)
	rows, err := db.Query(sql)
	if err != nil {
		panic(err)
	}
	var result = []body{}
	for rows.Next() {
		b := body{}
		err := rows.Scan(&b.Id, &b.Title, &b.CreationYear, &b.WhereToFind, &b.PictureAddress, &b.MuseumId, &b.AuthorId, &b.AuthorName, &b.MuseumName)
		if err != nil {
			fmt.Println(err)
			continue
		}
		result = append(result, b)
	}
	//c.HTML(http.StatusOK, "index.html", gin.H{"paintings": paintings})
	c.IndentedJSON(http.StatusOK, result)
}

func searchAuthor(c *gin.Context) {
	request := c.Param("request")
	query := fmt.Sprintf("select * from authors where name like '%s%%';", request)
	fmt.Println(query)
	rows, err := db.Query(query)
	if err != nil {
		panic(err)
	}
	var authors = []author{}
	for rows.Next() {
		a := author{}
		err := rows.Scan(&a.Id, &a.Name, &a.BirthYear, &a.DeathYear, &a.Biography)
		if err != nil {
			fmt.Println(err)
			continue
		}
		authors = append(authors, a)
	}
	//c.HTML(http.StatusOK, "index.html", gin.H{"paintings": paintings})
	c.IndentedJSON(http.StatusOK, authors)
}

func postPainting(c *gin.Context) {
	type body struct {
		Title          string `json:"title"`
		CreationYear   int    `json:"creation_year"`
		WhereToFind    string `json:"where_to_find"`
		PictureAddress string `json:"picture_address"`
		AuthorId       int    `json:"author_id"`
	}
	var b body
	if err := c.BindJSON(&b); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to read body"})
		return
	}

	user_id, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed create new painting"})
		return
	}

	museum_id, err := strconv.Atoi(c.Param("museum_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to create new painting"})
		return
	}

	has_permission := false
	query := fmt.Sprintf("select change_paintings from rights where user_id = %f and museum_id = %d", user_id, museum_id)
	err = db.QueryRow(query).Scan(&has_permission)
	if err != nil || !has_permission {
		c.JSON(http.StatusBadRequest, gin.H{"error": "you are not permitted to create paintings"})
		return
	}

	query = fmt.Sprintf("insert into paintings(title,creation_year,where_to_find,picture_address,author_id,museum_id) values('%s',%d,'%s','%s', %d, %d)",
		b.Title,
		b.CreationYear,
		b.WhereToFind,
		b.PictureAddress,
		b.AuthorId,
		museum_id)
	_, err = db.Exec(query)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to create new painting"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{})
}

func changePainting(c *gin.Context) {
	var newPainting painting
	if err := c.BindJSON(&newPainting); err != nil {
		fmt.Println(err)
		return
	}
	//fmt.Println("insert into paintings values (" + strconv.Itoa(newPainting.Id) + ", " + newPainting.Title + ", " + newPainting.Author + ")")
	//_, err := db.Exec("insert into paintings values (" + strconv.Itoa(newPainting.Id) + ", '" + newPainting.Title + "', '" + newPainting.Author + "')")
	/*
		if err != nil {
			fmt.Println(err.Error())
		}
	*/
	//paintings = append(paintings, newPainting)
	//c.IndentedJSON(http.StatusCreated, newPainting)
}

func deletePainting(c *gin.Context) {
	var newPainting painting
	if err := c.BindJSON(&newPainting); err != nil {
		fmt.Println(err)
		return
	}
	//fmt.Println("insert into paintings values (" + strconv.Itoa(newPainting.Id) + ", " + newPainting.Title + ", " + newPainting.Author + ")")
	//_, err := db.Exec("insert into paintings values (" + strconv.Itoa(newPainting.Id) + ", '" + newPainting.Title + "', '" + newPainting.Author + "')")
	/*
		if err != nil {
			fmt.Println(err.Error())
		}
	*/
	//paintings = append(paintings, newPainting)
	//c.IndentedJSON(http.StatusCreated, newPainting)
}

func login(c *gin.Context) {
	var body struct {
		Login    string
		Password string
	}
	if c.Bind(&body) != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to read body"})
		return
	}
	var password_hashed string
	var user_id int
	query := fmt.Sprintf("select id, password_hashed from users where login='%s'", body.Login)
	fmt.Println(query)
	err := db.QueryRow(query).Scan(&user_id, &password_hashed)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusBadRequest, gin.H{"error": "user with this login is not registered"})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": "internal database error"})
		return
	}
	err = bcrypt.CompareHashAndPassword([]byte(password_hashed), []byte(body.Password))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid password"})
		return
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{"sub": user_id, "exp": time.Now().Add(time.Hour).Unix()})
	var tokenString string
	tokenString, err = token.SignedString([]byte("top-secret"))

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "internal error"})
		return
	}

	// cookie version (modern?)
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie("Auth", tokenString, 3600, "", "", false, true)
	c.JSON(http.StatusOK, gin.H{})

	// sending token version
	//c.JSON(http.StatusOK, gin.H{"token": tokenString})
}

func getLoginInfo(c *gin.Context) {
	user_id, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to add favorite"})
		return
	}
	var user_name string
	query := fmt.Sprintf("select login from users where id = %f", user_id)
	fmt.Println(query)
	err := db.QueryRow(query).Scan(&user_name)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"user_name": user_name})
}

func register(c *gin.Context) {
	var body struct {
		Login    string
		Password string
	}

	if c.Bind(&body) != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to read body"})
		return
	}

	password_hashed, err := bcrypt.GenerateFromPassword([]byte(body.Password), 10)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to hash password"})
		return
	}

	query := fmt.Sprintf("Insert into users(login, password_hashed) values ('%s','%s')", body.Login, password_hashed)
	fmt.Println(query)
	_, err = db.Exec(query)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "couldn't register a user"})
	}

}

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

func postFavorite(c *gin.Context) {
	var body struct {
		PaintingId int `json:"painting_id"`
	}
	if c.Bind(&body) != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to read body"})
		return
	}

	user_id, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to add favorite"})
		return
	}

	query := fmt.Sprintf("insert into users_preferences values (%f, %d)", user_id, body.PaintingId)
	fmt.Println(query)
	_, err := db.Exec(query)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to add favorite"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{})
}

func deleteFavorite(c *gin.Context) {
	var body struct {
		PaintingId int `json:"painting_id"`
	}
	if c.Bind(&body) != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to read body"})
		return
	}

	user_id, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to remove favorite"})
		return
	}

	query := fmt.Sprintf("delete from users_preferences where user_id = %d and painting_id = %d ", user_id, body.PaintingId)
	_, err := db.Exec(query)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to remove favorite"})
		return
	}

	c.JSON(http.StatusNoContent, gin.H{})
}

func postRegisterMuseum(c *gin.Context) {
	var body struct {
		Name    string `json:"name"`
		WebPage string `json:"web_page"`
	}
	if c.Bind(&body) != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to read body"})
		return
	}

	user_id, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to create museum"})
		return
	}

	ctx := context.Background()
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to create museum"})
		return
	}

	query := fmt.Sprintf("insert into museums(name, web_page) values ('%s', '%s')", body.Name, body.WebPage)
	_, err = tx.ExecContext(ctx, query)
	if err != nil {
		tx.Rollback()
		fmt.Println("Transaction rollback!")
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to create museum"})
		return
	}
	var museum_id int
	query = fmt.Sprintf("select id from museums where name='%s'", body.Name)
	err = tx.QueryRowContext(ctx, query).Scan(&museum_id)
	if err != nil {
		tx.Rollback()
		fmt.Println("Transaction rollback!")
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to create museum"})
		return
	}
	query = fmt.Sprintf("insert into rights values (%f, %d, true, true)", user_id, museum_id)
	_, err = tx.ExecContext(ctx, query)
	if err != nil {
		tx.Rollback()
		fmt.Println("Transaction rollback!")
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to create museum"})
		return
	}
	err = tx.Commit()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to create museum"})
		return
	}
	c.JSON(http.StatusOK, gin.H{})
}

func getMuseumPaintings(c *gin.Context) {
	type body struct {
		Id             int    `json:"id"`
		Title          string `json:"title"`
		CreationYear   string `json:"creation_year"`
		WhereToFind    string `json:"where_to_find"`
		PictureAddress string `json:"picture_address"`
		MuseumId       int    `json:"museum_id"`
		AuthorId       int    `json:"author_id"`
		AuthorName     string `json:"author_name"`
	}

	museum_id, err := strconv.Atoi(c.Param("museum_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to remove rights"})
		return
	}

	query := fmt.Sprintf("select p.*, a.name from paintings as p join authors as a on p.author_id = a.id where p.museum_id = %d", museum_id)
	rows, err := db.Query(query)
	if err != nil {
		panic(err)
	}
	var result = []body{}
	for rows.Next() {
		b := body{}
		err := rows.Scan(&b.Id, &b.Title, &b.CreationYear, &b.WhereToFind, &b.PictureAddress, &b.MuseumId, &b.AuthorId, &b.AuthorName)
		if err != nil {
			fmt.Println(err)
			continue
		}
		result = append(result, b)
	}

	c.JSON(http.StatusOK, result)
}

func postUserRights(c *gin.Context) {
	res := CanChangeRights(c)
	if res == nil {
		return
	}
	query := fmt.Sprintf("insert into rights values (%d, %d, %t, %t)", res.UserId, res.MuseumId, res.GiveRights, res.ChangePaintings)
	_, err := db.Exec(query)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to give rights"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{})
}

func changeUserRights(c *gin.Context) {
	res := CanChangeRights(c)
	if res == nil {
		return
	}
	query := fmt.Sprintf("update rights set give_rights=%t, change_paintings=%t where user_id=%d and museum_id=%d", res.GiveRights, res.ChangePaintings, res.UserId, res.MuseumId)
	fmt.Println(query)
	_, err := db.Exec(query)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to change rights"})
		return
	}

	c.JSON(http.StatusOK, gin.H{})
}

func deleteUserRights(c *gin.Context) {
	var body struct {
		Login string `json:"login"`
	}
	if c.Bind(&body) != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to read body"})
		return
	}

	user_id, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to remove rights"})
		return
	}

	museum_id, err := strconv.Atoi(c.Param("museum_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to remove rights"})
		return
	}

	has_permission := false
	delete_rights_from_id := -1
	query := fmt.Sprintf("select give_rights from rights where user_id = %f and museum_id = %d", user_id, museum_id)
	err = db.QueryRow(query).Scan(&has_permission)
	if err != nil || !has_permission {
		c.JSON(http.StatusBadRequest, gin.H{"error": "you are not permitted to do this"})
		return
	}

	query = fmt.Sprintf("select id from users where login = '%s'", body.Login)
	err = db.QueryRow(query).Scan(&delete_rights_from_id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to remove rights"})
		return
	}
	if user_id == (float64)(delete_rights_from_id) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "you can't remove your own rights"})
		return
	}

	query = fmt.Sprintf("delete from rights where user_id=%d", delete_rights_from_id)
	_, err = db.Exec(query)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to remove rights"})
		return
	}

	c.JSON(http.StatusNoContent, gin.H{})
}

// A support function for postUserRights and setUserRights
func CanChangeRights(c *gin.Context) *rights {
	var body struct {
		Login           string `json:"login"`
		GiveRights      bool   `json:"give_rights"`
		ChangePaintings bool   `json:"change_paintings"`
	}
	if c.Bind(&body) != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to read body"})
		return nil
	}

	user_id, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to change rights"})
		return nil
	}

	museum_id, err := strconv.Atoi(c.Param("museum_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to change rights"})
		return nil
	}

	has_permission := false
	change_rights_user_id := -1
	query := fmt.Sprintf("select give_rights from rights where user_id = %f and museum_id = %d", user_id, museum_id)
	err = db.QueryRow(query).Scan(&has_permission)
	if err != nil || !has_permission {
		c.JSON(http.StatusBadRequest, gin.H{"error": "you are not permitted to do this"})
		return nil
	}
	query = fmt.Sprintf("select id from users where login = '%s'", body.Login)
	fmt.Println(query)
	err = db.QueryRow(query).Scan(&change_rights_user_id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to change rights"})
		return nil
	}
	if user_id == (float64)(change_rights_user_id) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "you can't change your own rights"})
		return nil
	}
	return &rights{change_rights_user_id, museum_id, body.GiveRights, body.ChangePaintings}
}
