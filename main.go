package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"golang.org/x/time/rate"
)

type weatherFetch struct {
	Location struct {
		Country string `json:"country"`
		Region  string `json:"region"`
	} `json:"location"`
	Current struct {
		Temperature float64 `json:"temp_c"`
		Feels       float64 `json:"feelslike_c"`

		Conditions struct {
			Condition string `json:"text"`
		} `json:"condition"`
	} `json:"current"`
}

func fetchFromApi(url string, country *string, region *string, temperature *float64, feels *float64, condition *string) {
	resp, err := http.Get(url)
	if err != nil {
		log.Fatalln(err)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}

	var weather = new(weatherFetch)
	response := json.Unmarshal(body, &weather)

	if response != nil {
		panic(response)
	}

	*country = weather.Location.Country
	*region = weather.Location.Region
	*temperature = weather.Current.Temperature
	*feels = weather.Current.Feels
	*condition = weather.Current.Conditions.Condition
}
func getCity(city *string) string {

	return *city

}

func RateLimiter() gin.HandlerFunc {
	limiter := rate.NewLimiter(1, 4)

	return func(c *gin.Context) {
		if limiter.Allow() {
			c.Next()
			return
		}

		c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
			"message": "Limit exceeded",
		})
	}
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	router := gin.Default()
	router.Use(RateLimiter())
	router.GET("/weather", func(c *gin.Context) {

		API_KEY := os.Getenv("WEATHER_API_KEY")

		city := c.DefaultQuery("city", "Guest")
		fmt.Println(getCity(&city))

		url := fmt.Sprintf("https://api.weatherapi.com/v1/current.json?key=%s&q=%s&aqi=no", API_KEY, getCity(&city))
		var country string
		var region string
		var temperature float64
		var feels float64
		var condition string

		fetchFromApi(url, &country, &region, &temperature, &feels, &condition)
		fmt.Println(country, region, temperature, feels, condition)

		c.JSON(http.StatusOK, gin.H{
			"country":     country,
			"region":      region,
			"temperature": temperature,
			"feels_like":  feels,
			"condition":   condition,
		})

	})

	router.Run("localhost:8080")

}
