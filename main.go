package main

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
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
	PictureAddress string        `json:"picture_address" binding:"required"`
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
	Name     string `gorm:"unique" binding:"required"`
	WebPage  string `binding:"required"`
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
	UserID          int    `gorm:"primaryKey" json:"user_id" binding:"required"`
	User            User   `gorm:"foreignKey:user_id" binding:"-"`
	MuseumID        int    `gorm:"primaryKey"`
	Museum          Museum `gorm:"foreignKey:museum_id" binding:"-"`
	GiveRights      *bool  `json:"give_rights" binding:"required"`
	ChangePaintings *bool  `json:"change_paintings" binding:"required"`
}

var db *gorm.DB

func main() {

	//=====================================SETTING UP THE DATABASE GORM===========================================
	connStr := "host=localhost user=postgres password=pass dbname=Paintings_Web_App port=5431 sslmode=disable"
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

	//====================================SETTING UP THE ROUTER===================================================
	router := gin.Default()
	router.Static("/assets", "./assets")
	router.LoadHTMLGlob("templates/*.html")

	router.GET("/", getLanding)

	router.GET("/paintings_:request", searchPainting)
	router.GET("/authors_:request", searchAuthor)

	router.POST("/login", login)
	router.POST("/register", register)

	router.POST("/register_museum", requireAuth, postRegisterMuseum)

	//TODO: check how dates are represented here in json
	router.POST("/museum:museum_id/create_painting", requireAuth, postPainting)
	//TODO: check how dates are represented here in json
	router.GET("/museum:museum_id/all_paintings", getMuseumPaintings)

	//router.GET("/museum:museum_id/rights", requireAuth, getAllUsersRights)
	router.POST("/museum:museum_id/rights", requireAuth, postUserRights)
	router.PUT("/museum:museum_id/rights", requireAuth, changeUserRights)
	router.DELETE("/museum:museum_id/rights", requireAuth, deleteUserRights)

	//TODO: check how dates are represented here in json
	router.PUT("/painting:painting_id/change_painting", requireAuth, changePainting)
	router.DELETE("/painting:painting_id/delete_painting", requireAuth, deletePainting)

	router.GET("/favorite", requireAuth, getFavorites)
	router.POST("/favorite", requireAuth, postFavorite)
	router.DELETE("/favorite", requireAuth, deleteFavorite)

	router.GET("/login_info", requireAuth, getLoginInfo)

	router.Run("localhost:8080")

	/*
		var birth_year sql.NullInt32
		var death_year sql.NullInt32
		birth_year.Int32 = 2000
		birth_year.Valid = true
		death_year.Int32 = 0
		death_year.Valid = true
		testAuthor := Author{Name: "Bill", BirthYear: birth_year, DeathYear: death_year, Biography: ""}
		db.Create(&testAuthor)
	*/
	//db.Create(&testMuseum)
	//user_preference := UserPreference{UserID: 1, PaintingID: 1}
	/*
		var creation_year sql.NullInt32
		creation_year.Int32 = 2010
		creation_year.Valid = true
		painting := Painting{Title: "Hello", CreationYear: creation_year, WhereToFind: "", PictureAddress: "", AuthorID: 1, MuseumID: 2}
		db.Create(&painting)
	*/

	/*right := Right{UserID: 1, MuseumID: 3}
	db.Create(&right)*/

}

func getLanding(c *gin.Context) {
	c.HTML(http.StatusOK, "index.html", gin.H{})
}

func searchPainting(c *gin.Context) {
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
	var results []Result = []Result{}
	request := c.Param("request")
	//fmt.Println("request:" + request)
	err := db.Raw("select p.*, a.name as author_name, m.name as museum_name from ((select * from paintings where LOWER(title) like LOWER(?)) as p join authors as a on p.author_id=a.id) join museums as m on p.museum_id = m.id", "%%"+request+"%%").Scan(&results).Error
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "something went wrong"})
	}
	//c.HTML(http.StatusOK, "index.html", gin.H{"paintings": paintings})
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
	db.Where("lower(name) like lower(?)", "%%"+request+"%%").Find(&authors)
	//c.HTML(http.StatusOK, "index.html", gin.H{"paintings": paintings})
	c.IndentedJSON(http.StatusOK, authors)
}

func postPainting(c *gin.Context) {
	var painting Painting
	var err error
	if err := c.BindJSON(&painting); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to read body"})
		return
	}

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
	if err != nil || !(*userRight.ChangePaintings) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "you are not permitted to add new paintings to the collection of this museum"})
		return
	}

	err = db.First(&Author{}, painting.AuthorID).Error
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "author id is not correct"})
		return
	}

	err = db.Create(&painting).Error
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to create new painting"})
		return
	}

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
	if err != nil || !(*userRight.ChangePaintings) {
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
	/*
		var newPainting Painting
		if err := c.BindJSON(&newPainting); err != nil {
			fmt.Println(err)
			return
		}
		//fmt.Println("insert into paintings values (" + strconv.Itoa(newPainting.Id) + ", " + newPainting.Title + ", " + newPainting.Author + ")")
		//_, err := db.Exec("insert into paintings values (" + strconv.Itoa(newPainting.Id) + ", '" + newPainting.Title + "', '" + newPainting.Author + "')")
		//paintings = append(paintings, newPainting)
		//c.IndentedJSON(http.StatusCreated, newPainting)
	*/
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
	var user_name string
	query := fmt.Sprintf("select login from users where id = %f", user_id)
	fmt.Println(query)
	var u User
	err := db.First(&u).Error
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"user_name": user_name})
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
	}

	c.JSON(http.StatusOK, gin.H{})
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
	var res Result
	//need to be done almost with json
	userId, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to add favorite"})
		return
	}
	//then return all favorites
	err := db.Raw("select p.*, a.name as author_name, m.name as museum_name from (((select * from paintings where id in (select painting_id from user_preferences where user_id = ?)) as p join authors as a on p.author_id=a.id) join museum as m on p.museum_id = m.id)", userId).Scan(&res).Error
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

func postRegisterMuseum(c *gin.Context) {
	var newMuseum Museum = Museum{}
	var newRight Right = Right{}
	if c.Bind(&newMuseum) != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to read body"})
		return
	}
	newMuseum.Verified = false
	userId, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to create museum"})
		return
	}

	err := db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&newMuseum).Error; err != nil {
			return err
		}
		newRight.MuseumID = newMuseum.ID
		newRight.UserID = int(userId.(float64))
		*newRight.GiveRights = true
		*newRight.ChangePaintings = true
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to interpret museum id"})
		return
	}
	//reason for type change of painting: we want to interpret years as strings
	var paintings []Painting
	err = db.Where("museum_id = ?", museum_id).Find(&paintings).Error
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to find paintings"})
	}
	c.JSON(http.StatusOK, paintings)
}

func postUserRights(c *gin.Context) {
	var err error
	var newRight Right = Right{}
	newRight.MuseumID, err = strconv.Atoi(c.Param("museum_id"))
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
		return
	}
	err = c.Bind(&newRight)
	if err != nil {
		fmt.Println(err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to read body"})
		return
	}
	if !CanChangeRights(c, &newRight) {
		return
	}
	err = db.Save(&newRight).Error
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to change rights"})
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
		return
	}
	err = c.Bind(&body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to read body"})
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
	if err != nil || !(*permitted.ChangePaintings) {
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
