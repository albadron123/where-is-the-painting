package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"path"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Painting struct {
	ID             int           `json:"id"`
	Title          string        `json:"title" binding:"required"`
	CreationYear   sql.NullInt32 `json:"creation_year" gorm:"check:(creation_year>=0)"`
	WhereToFind    string        `json:"where_to_find" binding:"required"`
	PictureAddress string        `binding:"-"`
	AuthorID       int           `json:"author_id" binding:"required"`
	Author         Author        `gorm:"foreignKey:author_id" binding:"-"`
	MuseumID       int
	Museum         Museum `gorm:"foreignKey:museum_id" binding:"-"`
}

type User struct {
	ID             int
	Login          string `gorm:"unique"`
	PasswordHashed string
}

type Museum struct {
	ID       int
	Name     string `json:"name" gorm:"unique" binding:"required"`
	WebPage  string `json:"webpage" binding:"required"`
	Verified bool   `gorm:"default:false"`
}

// TODO: try to set constraints against the current year in GORM (if possible)
type Author struct {
	ID        int
	Name      string        `gorm:"not null"`
	BirthYear sql.NullInt32 `gorm:"check:(birth_year>=0)"`
	DeathYear sql.NullInt32 `gorm:"check:(death_year>birth_year)"`
	Biography string
}

type UserPreference struct {
	UserID     int      `gorm:"primaryKey"`
	User       User     `gorm:"foreignKey:user_id"`
	PaintingID int      `gorm:"primaryKey"`
	Painting   Painting `gorm:"foreignKey:painting_id"`
}

type Right struct {
	UserID   int    `gorm:"primaryKey" json:"user_id" binding:"required"`
	User     User   `gorm:"foreignKey:user_id" binding:"-"`
	MuseumID int    `gorm:"primaryKey" binding:"-"`
	Museum   Museum `gorm:"foreignKey:museum_id" binding:"-"`
	IsAdmin  bool   `json:"is_admin"`
}

var db *gorm.DB

func main() {

	fmt.Println("5432PORT")
	//=====================================SETTING UP THE DATABASE GORM===========================================
	connStr := "host=pg user=postgres password=pass dbname=Paintings_Web_App port=5432 sslmode=disable"
	var err error
	db, err = gorm.Open(postgres.New(postgres.Config{
		DSN:                  connStr,
		PreferSimpleProtocol: true,
	}), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	fmt.Println("GORM_DB", db)

	//setting auto-migrations for tables in db
	//note: when connecting to the db through the db connection with database/sql and then connecting it to gorm
	//		auto-migration gives strange errors saying something about pq.
	db.AutoMigrate(&User{}, &Author{}, &Painting{}, &Museum{}, &UserPreference{}, &Right{})

	//====================================CREATING DATA===========================================================

	addAuthorsIntoDB(20)
	addPaintingsIntoDB(100, 2, 1)

	//====================================SETTING UP THE ROUTER===================================================
	router := gin.Default()
	router.Static("/assets", "./assets")
	router.Static("/css", "./css")
	router.Static("/js", "./js")
	router.LoadHTMLGlob("./templates/*")

	router.GET("/", getMain_PAGE)

	router.GET("/debug", get_DEBUG_PAGE)
	router.POST("/debug", post_DEBUG)

	router.GET("/success", success_PAGE)

	router.GET("/museum:museum_id/search_users_:request", searchNewUsers)

	router.GET("/paintings_:request", searchPainting)
	router.GET("/login_paintings_:request", requireAuth, searchPaintingsLoggedIn)
	router.GET("/authors_:request", searchAuthor)

	router.POST("/login", login)
	router.POST("/register", register)

	router.GET("/register_author", requireAuthPAGE, getRegisterAuthor_PAGE)
	router.POST("/register_author", requireAuth, postRegisterAuthor)

	router.GET("/register_museum", requireAuthPAGE, getRegisterMuseum_PAGE)
	router.POST("/register_museum", requireAuth, postRegisterMuseum)

	router.GET("/museum:museum_id", requireAuthPAGE, getMuseum_PAGE)

	router.GET("/museum:museum_id/register_painting", requireAuthPAGE, registerPainting_PAGE)
	router.POST("/museum:museum_id/register_painting", requireAuth, postPainting)
	router.GET("/museum:museum_id/paintings_:request/page:page_id", getMuseumPaintings)

	//router.GET("/museum:museum_id/rights", requireAuth, getAllUsersRights)
	//COMMENT: for now is deleted as museum users rights are aquired and rendered by the page route that returns the rendered template.
	router.POST("/museum:museum_id/rights", requireAuth, postUserRights)
	router.PUT("/museum:museum_id/rights", requireAuth, changeUserRights)
	router.DELETE("/museum:museum_id/rights", requireAuth, deleteUserRights)

	router.PUT("/painting:painting_id/change_painting", requireAuth, changePainting)
	router.DELETE("/painting:painting_id/delete_painting", requireAuth, deletePainting)

	router.GET("/favorite", requireAuth, getFavorites)
	router.POST("/favorite", requireAuth, postFavorite)
	router.DELETE("/favorite", requireAuth, deleteFavorite)

	router.GET("/login_info", requireAuth, getLoginInfo)

	router.Run("0.0.0.0:8080")
}

func addPaintingsIntoDB(count int, authorId int, museumId int) {
	for i := 0; i < count; i++ {
		p := Painting{}
		p.Title = generateStr(20)
		p.AuthorID = authorId
		p.MuseumID = museumId
		p.CreationYear = sql.NullInt32{Int32: int32(1900 + rand.Intn(100)), Valid: true}
		p.WhereToFind = generateStr(10)
		db.Create(&p)
	}
}

func addAuthorsIntoDB(count int) {
	for i := 0; i < count; i++ {
		a := Author{}
		a.Name = generateStr(20)
		birth := int32(1900 + rand.Intn(100))
		a.BirthYear = sql.NullInt32{Int32: birth, Valid: true}
		a.DeathYear = sql.NullInt32{Int32: birth + int32(rand.Intn(20)), Valid: true}
		a.Biography = generateStr(50)
		db.Create(&a)
	}
}

func searchNewUsers(c *gin.Context) {
	type In struct {
		MuseumId int    `uri:"museum_id" binding:"required"`
		Request  string `uri:"request" binding:"required"`
	}
	type Result struct {
		ID    int    `json:"id"`
		Login string `json:"login" gorm:"unique"`
	}

	in := In{}
	if err := c.ShouldBindUri(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to read uri"})
		fmt.Print(err.Error())
		return
	}

	var results []Result = []Result{}
	err := db.Raw(
		`SELECT id, login 
		 FROM users 
		 WHERE 
		 	id NOT IN (SELECT user_id FROM rights WHERE museum_id = ?) AND
			LOWER(login) like LOWER(?) 
		 ORDER BY position (lower(?) in lower(login))
		 LIMIT 5`,
		in.MuseumId, "%%"+in.Request+"%%", in.Request).Scan(&results).Error
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "something went wrong"})
		return
	}

	c.IndentedJSON(http.StatusOK, results)
}

func searchPainting(c *gin.Context) {
	type Result struct {
		ID             int    `json:"id"`
		Title          string `json:"title"`
		CreationYear   string `json:"creation_year"`
		WhereToFind    string `json:"where_to_find"`
		PictureAddress string `json:"picture_address"`
		MuseumId       int    `json:"museum_id"`
		AuthorId       int    `json:"author_id"`
		AuthorName     string `json:"author_name"`
		MuseumName     string `json:"museum_name"`
	}
	var results []Result = []Result{}
	request := c.Param("request")
	//fmt.Println("request:" + request)
	err := db.Raw("select p.*, a.name as author_name, m.name as museum_name from ((select * from paintings where LOWER(title) like LOWER(?)) as p join authors as a on p.author_id=a.id) join museums as m on p.museum_id = m.id order by position (lower(?) in lower(p.title)) limit 100", "%%"+request+"%%", request).Scan(&results).Error
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "something went wrong"})
	}
	//c.HTML(http.StatusOK, "index.html", gin.H{"paintings": paintings})
	c.IndentedJSON(http.StatusOK, results)
}

func searchPaintingsLoggedIn(c *gin.Context) {
	type Result struct {
		Liked          int    `json:"liked"`
		ID             int    `json:"id"`
		Title          string `json:"title"`
		CreationYear   string `json:"creation_year"`
		WhereToFind    string `json:"where_to_find"`
		PictureAddress string `json:"picture_address"`
		MuseumId       int    `json:"museum_id"`
		AuthorId       int    `json:"author_id"`
		AuthorName     string `json:"author_name"`
		MuseumName     string `json:"museum_name"`
	}

	userId, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to search for a painting"})
		return
	}

	request := c.Param("request")

	fmt.Println(request)
	fmt.Println(userId)
	var results []Result = []Result{}
	fmt.Println("!")

	err := db.Exec(`
		create or replace temp view general_search as
		select p.*, a.name as author_name, m.name as museum_name from
		((select * from paintings where LOWER(title) like LOWER(?)) as p join authors as a on p.author_id=a.id) join museums as m on p.museum_id = m.id 
		order by position (lower(?) in lower(p.title)) 
		limit 100;
		`, "%%"+request+"%%", request).Error
	if err != nil {
		fmt.Println(err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to create view"})
		return
	}

	err = db.Exec(`
		create or replace view liked_ones as
		select painting_id from user_preferences where user_id = ? and painting_id in (select id from general_search);
		`, userId).Error
	if err != nil {
		fmt.Println(err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to create view"})
		return
	}

	err = db.Raw(`
		select sub.liked, general_search.* from
		((select painting_id, 1 as "liked" from liked_ones) union
		(select id, 0 as "liked" from general_search where id not in (select painting_id from liked_ones))) as sub join general_search on
		general_search.id = sub.painting_id;
		`).Scan(&results).Error
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "something went wrong"})
		return
	}

	c.IndentedJSON(http.StatusOK, results)
}

func searchAuthor(c *gin.Context) {
	type Author struct {
		ID        int    `json:"id"`
		Name      string `json:"name"`
		BirthYear string `json:"birth_year"`
		DeathYear string `json:"death_year"`
		Biography string `json:"biography"`
	}

	request := c.Param("request")
	var authors = []Author{}
	err := db.Raw("select * from authors where lower(name) like lower(?) order by position (lower(?) in lower(name)) limit 5", "%%"+request+"%%", request).Find(&authors).Error
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to find paintings"})
	}
	//c.HTML(http.StatusOK, "index.html", gin.H{"paintings": paintings})
	c.IndentedJSON(http.StatusOK, authors)
}

func postPainting(c *gin.Context) {
	var err error
	var painting Painting

	userId, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed create new painting"})
		return
	}

	painting.MuseumID, err = strconv.Atoi(c.Param("museum_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to create new painting"})
		return
	}

	err = db.First(&Museum{}, painting.MuseumID).Error
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "museum id is not correct"})
		return
	}

	var userRight Right
	err = db.Where("user_id = ? and museum_id = ?", userId, painting.MuseumID).First(&userRight).Error
	//there is no rights in the table (or other mistakes)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "you are not permitted to add new paintings to the collection of this museum"})
		return
	}

	in := []byte(c.Request.FormValue("json"))
	file, _ := c.FormFile("file")

	err = json.Unmarshal(in, &painting)
	if err != nil {
		fmt.Println(err.Error())
		c.JSON(http.StatusBadRequest, gin.H{})
		return
	}

	err = db.First(&Author{}, painting.AuthorID).Error
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "author id is not correct"})
		return
	}

	err = db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&painting).Error; err != nil {
			return err
		}

		fileExt := path.Ext(file.Filename)

		painting.PictureAddress = strconv.Itoa(painting.ID) + fileExt
		err = tx.Save(&painting).Error
		if err != nil {
			return err
		}

		err = c.SaveUploadedFile(file, "./assets/"+painting.PictureAddress)
		if err != nil {
			fmt.Println(err.Error())
			fmt.Println("error")
			return err
		}

		return nil
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to register new painting"})
		return
	}

	fmt.Println("Hof!")
	c.JSON(http.StatusCreated, gin.H{})
}

func changePainting(c *gin.Context) {
	var painting Painting
	var err error
	if err := c.BindJSON(&painting); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to read body"})
		return
	}

	userId, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed change painting"})
		return
	}

	painting.MuseumID, err = strconv.Atoi(c.Param("museum_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to change painting"})
		return
	}

	err = db.First(&Museum{}, painting.MuseumID).Error
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "museum id is not correct"})
		return
	}

	var userRight Right
	err = db.Where("user_id = ? and museum_id = ?", userId, painting.MuseumID).First(&userRight).Error
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "you are not permitted to change paintings in the collection of this museum"})
		return
	}

	painting.ID, err = strconv.Atoi(c.Param("painting_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to change painting"})
	}

	err = db.First(&Author{}, painting.AuthorID).Error
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "author id is not correct"})
		return
	}

	err = db.Save(&painting).Error
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to change painting"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{})
}

func deletePainting(c *gin.Context) {
	paintingId, err := strconv.Atoi(c.Param("painting_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to read uri"})
		return
	}

	err = db.Transaction(func(tx *gorm.DB) error {
		var p Painting
		err := db.Where("id = ?", paintingId).First(&p).Error
		if err != nil {
			fmt.Println(err.Error())
			return err
		}
		err = os.Remove("assets/" + p.PictureAddress)
		if err != nil {
			if !errors.Is(err, os.ErrNotExist) {
				fmt.Println(err.Error())
				return err
			}
		}

		err = db.Where("painting_id = ?", paintingId).Delete(&UserPreference{}).Error
		if err != nil {
			fmt.Println(err.Error())
			return err
		}

		err = db.Delete(p).Error
		if err != nil {
			fmt.Println(err.Error())
			return err
		}
		return nil
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to delete painting"})
		return
	}
	c.JSON(http.StatusNoContent, gin.H{})
}

func login(c *gin.Context) {
	var body struct {
		Login    string `json:"login" binding:"required"`
		Password string `json:"password" binding:"required"`
	}
	if c.Bind(&body) != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to read body"})
		return
	}

	var user User
	err := db.Where("login = ?", body.Login).First(&user).Error
	if err != nil {
		fmt.Println(err)
		if errors.Is(err, gorm.ErrRecordNotFound) {

			c.JSON(http.StatusBadRequest, gin.H{"error": "user with this login is not registered"})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": "internal database error"})
		return
	}
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHashed), []byte(body.Password))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid password"})
		return
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{"sub": user.ID, "exp": time.Now().Add(time.Hour).Unix()})
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
	fmt.Println("user_id:", user_id)
	//var user_id,user_name string
	//query := fmt.Sprintf("select login from users where id = %f", user_id)
	//fmt.Println(query)
	var u User
	err := db.First(&u, user_id).Error
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user not found"})
		return
	}
	var museums []Museum
	err = db.Raw("SELECT id, name FROM museums WHERE id IN (SELECT museum_id FROM rights WHERE user_id = ?)", user_id).Find(&museums).Error
	if err != nil {
		fmt.Println(err.Error())
		if errors.Is(err, gorm.ErrRecordNotFound) {
			fmt.Println("nothing found")
			museums = []Museum{}
		} else {
			c.JSON(http.StatusBadRequest, gin.H{})
			return
		}
	}
	c.JSON(http.StatusOK, gin.H{"login": u.Login, "museums": museums})
}

func register(c *gin.Context) {

	var body struct {
		Login    string `json:"login" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	if c.Bind(&body) != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to read body"})
		return
	}

	passwordHashed, err := bcrypt.GenerateFromPassword([]byte(body.Password), 10)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to hash password"})
		return
	}

	newUser := User{Login: body.Login, PasswordHashed: string(passwordHashed)}
	err = db.Create(&newUser).Error
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "couldn't register a user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{})
}

func getFavorites(c *gin.Context) {
	type Result struct {
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
	var res []Result
	//need to be done almost with json
	userId, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to add favorite"})
		return
	}
	//then return all favorites
	err := db.Raw("select p.*, a.name as author_name, m.name as museum_name from (((select * from paintings where id in (select painting_id from user_preferences where user_id = ?)) as p join authors as a on p.author_id=a.id) join museums as m on p.museum_id = m.id)", userId).Scan(&res).Error
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed get favorites"})
		return
	}
	c.IndentedJSON(http.StatusOK, res)
}

func postFavorite(c *gin.Context) {

	var body struct {
		PaintingId int `json:"painting_id" binding:"required"`
	}
	if c.Bind(&body) != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to read body"})
		return
	}

	userId, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to add favorite"})
		return
	}

	var newUserPreference UserPreference
	newUserPreference.UserID = int(userId.(float64))
	newUserPreference.PaintingID = body.PaintingId
	err := db.Create(&newUserPreference).Error
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to add favorite"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{})

}

func deleteFavorite(c *gin.Context) {

	var body struct {
		PaintingId int `json:"painting_id" binding:"required"`
	}
	if c.Bind(&body) != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to read body"})
		return
	}

	userId, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to remove favorite"})
		return
	}

	err := db.Where("user_id = ? and painting_id = ?", userId, body.PaintingId).Delete(&UserPreference{}).Error
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to remove favorite"})
		return
	}

	c.JSON(http.StatusNoContent, gin.H{})
}

func postRegisterAuthor(c *gin.Context) {
	var newAuthor Author = Author{}
	if c.Bind(&newAuthor) != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to read body"})
		return
	}

	if err := db.Create(&newAuthor).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to register new author"})
		return
	}

	c.JSON(http.StatusOK, gin.H{})
}

func postRegisterMuseum(c *gin.Context) {
	var newMuseum Museum = Museum{}
	newRight := Right{}
	err := c.Bind(&newMuseum)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to read body"})
		fmt.Println(err.Error())
		return
	}
	newMuseum.Verified = false
	userId, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to create museum"})
		fmt.Println("user doesnt exist")
		return
	}

	err = db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&newMuseum).Error; err != nil {
			return err
		}
		newRight.MuseumID = newMuseum.ID
		newRight.UserID = int(userId.(float64))
		newRight.IsAdmin = true
		if err := tx.Create(&newRight).Error; err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to create museum"})
		return
	}
	c.JSON(http.StatusOK, gin.H{})
}

func getMuseumPaintings(c *gin.Context) {
	type Painting struct {
		ID             int    `json:"id"`
		Title          string `json:"title"`
		CreationYear   string `json:"creation_year"`
		WhereToFind    string `json:"where_to_find"`
		PictureAddress string `json:"picture_address"`
		MuseumId       int    `json:"museum_id"`
		AuthorId       int    `json:"author_id"`
		AuthorName     string `json:"author_name"`
	}
	type In struct {
		MuseumId int    `uri:"museum_id" binding:"required"`
		Req      string `uri:"request"   binding:"required"`
		PageId   int    `uri:"page_id"`
	}
	//router.GET("/museum:museum_id/paintings_:request/:page_id", getMuseumPaintings)
	in := In{}
	/*
		fmt.Println(c.Param("page_id"))
		fmt.Println(c.Param("museum_id"))
		fmt.Println(c.Param("request"))
	*/
	if err := c.ShouldBindUri(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to read uri"})
		fmt.Print(err.Error())
		return
	}
	pageSize := 10
	/*
		museum_id, err := strconv.Atoi(c.Param("museum_id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "failed to interpret museum id"})
			return
		}
	*/
	//reason for type change of painting: we want to interpret years as strings
	var paintings []Painting
	err := db.Raw(`select p.*, a.name as author_name from 
					((select * from paintings 
						where museum_id = ? and LOWER(title) like LOWER(?)) as p join authors as a on p.author_id=a.id) 
					order by position (lower(?) in lower(p.title)) offset ? limit ? `,
		in.MuseumId, "%%"+in.Req+"%%", in.Req, pageSize*in.PageId, pageSize).Find(&paintings).Error
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to find paintings"})
		return
	}
	c.JSON(http.StatusOK, paintings)
}

func postUserRights(c *gin.Context) {
	var err error
	var newRight Right = Right{}
	newRight.MuseumID, err = strconv.Atoi(c.Param("museum_id"))
	println(newRight.MuseumID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid museum id"})
		return
	}
	err = c.Bind(&newRight)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to read body"})
		return
	}
	if !CanChangeRights(c, &newRight) {
		return
	}
	err = db.Create(&newRight).Error
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to give rights"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{})
}

func changeUserRights(c *gin.Context) {
	var err error
	var newRight Right = Right{}
	newRight.MuseumID, err = strconv.Atoi(c.Param("museum_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid museum id"})
		fmt.Println(err.Error())
		return
	}
	err = c.Bind(&newRight)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to read body"})
		fmt.Println(err.Error())
		return
	}
	fmt.Println("user", newRight.UserID)
	if !CanChangeRights(c, &newRight) {
		fmt.Println("No permission")
		return
	}
	err = db.Save(&newRight).Error
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to change rights"})
		fmt.Println(err.Error())
		return
	}
	c.JSON(http.StatusCreated, gin.H{})
}

func deleteUserRights(c *gin.Context) {
	type Body struct {
		UserId int `json:"user_id" binding:"required"`
	}
	var body Body = Body{}
	var err error
	var right Right = Right{}
	right.MuseumID, err = strconv.Atoi(c.Param("museum_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid museum id"})
		fmt.Println("1")
		return
	}
	err = c.Bind(&body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to read body"})
		fmt.Println(err.Error())
		return
	}
	right.UserID = body.UserId
	if !CanChangeRights(c, &right) {
		return
	}
	err = db.Where("user_id = ? and museum_id = ?", right.UserID, right.MuseumID).Delete(&Right{}).Error
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed remove right"})
		return
	}
	c.JSON(http.StatusNoContent, gin.H{})
}

// A support function for postUserRights and setUserRights
func CanChangeRights(c *gin.Context, right *Right) bool {

	userId, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return false
	}

	var permitted Right
	err := db.Where("user_id = ? and museum_id = ?", userId, right.MuseumID).First(&permitted).Error
	if err != nil || !(permitted.IsAdmin) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user has no permission to change rights"})
		return false
	}
	var user User
	err = db.Where("id = ?", right.UserID).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {

			c.JSON(http.StatusBadRequest, gin.H{"error": "couldn't find a user"})

		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": "failed to change rights"})
		}
		return false
	}
	if user.ID == int(userId.(float64)) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "you can't change your own rights"})
		return false
	}
	return true
}
