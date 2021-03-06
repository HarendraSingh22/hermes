package utils

import (
	"encoding/json"
	"fmt"
    "math/rand"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/spf13/viper"
    "gopkg.in/redis.v5"
)

func GetDBConn() *gorm.DB {

	dbUser := viper.GetString("database.user")
	dbName := viper.GetString("database.name")
	dbPassword := viper.GetString("database.password")

	databaseCredentials := fmt.Sprintf("user=%s dbname=%s sslmode=disable password=%s", dbUser, dbName, dbPassword)
	db, err := gorm.Open("postgres", databaseCredentials)

	if err != nil {
		panic("[ERROR] Database connection failed")
	}

	return db
}

func GetRedisClient() *redis.Client {

    redisClient := redis.NewClient(&redis.Options{
        Addr:     "localhost:6379",
        Password: "", // no password set
        DB:       0,  // use default DB
    })

    return redisClient

}

func ParseJSON(form io.Reader) map[string]string {

	data := make(map[string]string)
	err := json.NewDecoder(form).Decode(&data)

	if err != nil {
		// TODO
		panic("Improve Error Message")
	}

	return data
}

func GenRandStr(n int) string {
    var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

    b := make([]rune, n)
    for i := range b {
        b[i] = letterRunes[rand.Intn(len(letterRunes))]
    }
    return string(b)
}

func SuccessResponse() map[string]string {
	response := make(map[string]string)
	response["status"] = "OK"
	return response
}

func RequestWebkiosk(username, password string) (bool, string) {
	// password should be url encoded
	password = url.QueryEscape(password)

	webkioskURL := "https://webkiosk.jiit.ac.in/CommonFiles/UserAction.jsp?txtInst=Institute&InstCode=JIIT&txtuType=Member%%20Type%%20&UserType=S&txtCode=Enrollment%%20No&MemberCode=%s&txtPin=Password%%2FPin&Password=%s&BTNSubmit=Submit"

	reqURL := fmt.Sprintf(webkioskURL, username, password)

	resp, err := http.Get(reqURL)
	if err != nil {
		log.Fatal(err)
	}

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	if strings.Contains(string(respBody), "PersonalFiles/ShowAlertMessageSTUD.jsp") || strings.Contains(string(respBody), "StudentPageFinal.jsp") {
		return true, "Valid Credentials"
	}
	if strings.Contains(string(respBody), "Invalid Password") {
		return false, "Invalid Password"
	}
	if strings.Contains(string(respBody), "Login Account Locked") {
		return false, "Account Locked"
	}
	if strings.Contains(string(respBody), "Wrong Member Type or Code") || strings.Contains(string(respBody), "correct institute name and enrollment") {
		return false, "Invalid Enrollment Number"
	}
	return false, "Unknown Error"
}
