package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/line/line-bot-sdk-go/linebot"
	"github.com/machinebox/graphql"
)

func main() {
	port := os.Getenv("PORT")

	if port == "" {
		log.Fatal("$PORT must be set")
	}

	router := gin.New()
	router.Use(gin.Logger())

	router.POST("/hook", func(c *gin.Context) {
		client := &http.Client{Timeout: time.Duration(15 * time.Second)}

		channelSecret := os.Getenv("LINE_CHANNEL_SECRET")
		channelAccessToken := os.Getenv("LINE_CHANNEL_ACCESSTOKEN")
		bot, err := linebot.New(channelSecret, channelAccessToken, linebot.WithHTTPClient(client))
		if err != nil {
			fmt.Println(err)
			return
		}
		received, err := bot.ParseRequest(c.Request)

		for _, event := range received {
			if event.Type == linebot.EventTypeMessage {
				switch message := event.Message.(type) {
				case *linebot.LocationMessage:
					nearStation, err := getClosestStation(message.Latitude, message.Longitude)
					if err != nil {
						postMessage := linebot.NewTextMessage("ERROR: " + err.Error())
						if _, err = bot.ReplyMessage(event.ReplyToken, postMessage).Do(); err != nil {
							log.Print(err)
						}
					}

					postMessage := linebot.NewTextMessage("最寄り駅は、" + nearStation.Name + "駅です。")
					if _, err = bot.ReplyMessage(event.ReplyToken, postMessage).Do(); err != nil {
						log.Print(err)
					}
				}
			}
		}
	})

	router.Run(":" + port)
}

// ClosestStationResponse is GraphQL query response
type ClosestStationResponse struct {
	Address string
	Name    string
	Lines   []ClosestStationLine
}

// ClosestStationLine is GraphQL query response
type ClosestStationLine struct {
	Name string
}

func getClosestStation(lat float64, lon float64) (station *ClosestStationResponse, err error) {
	latStr := strconv.FormatFloat(lat, 'f', 10, 64)
	lonStr := strconv.FormatFloat(lon, 'f', 10, 64)
	client := graphql.NewClient("https://sapi.tinykitten.me/")
	req := graphql.NewRequest(`
    query ($latitude: Float!, $longitude: Float!) {
		stationByCoords(latitude: $latitude, longitude: $longitude) {
		  name
		  address
		  lines {
			name
		  }
		}
	  }
	`)
	req.Var("latitude", latStr)
	req.Var("longitude", lonStr)

	ctx := context.Background()

	// run it and capture the response
	var result ClosestStationResponse
	if err := client.Run(ctx, req, &result); err != nil {
		log.Fatal(err)
	}

	return &result, nil
}
