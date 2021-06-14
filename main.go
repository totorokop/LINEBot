package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/TinyKitten/LINEBot/models"
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
			log.Println(err)
			return
		}
		received, err := bot.ParseRequest(c.Request)
		if err != nil {
			log.Println(err)
			return
		}

		for _, event := range received {
			if event.Type == linebot.EventTypeMessage {
				switch message := event.Message.(type) {
				case *linebot.LocationMessage:
					closestStation, err := getClosestStation(message.Latitude, message.Longitude)
					if err != nil {
						postMessage := linebot.NewTextMessage("ERROR: " + err.Error())
						if _, err = bot.ReplyMessage(event.ReplyToken, postMessage).Do(); err != nil {
							log.Print(err)
						}
					}

					stationName := closestStation.StationByCoords.Name
					stationAddr := closestStation.StationByCoords.Address
					distance := closestStation.StationByCoords.Distance
					lines := closestStation.StationByCoords.Lines
					linesStr := ""
					for i, line := range lines {
						if i != len(lines)-1 {
							linesStr += fmt.Sprintf("%s\n", line.Name)
						} else {
							linesStr += line.Name
						}
					}
					postMessage := linebot.NewTextMessage(
						fmt.Sprintf(
							"最寄り駅情報\n%s駅\n%s\n駅からの直線距離:%.1fm\n\n利用可能路線:\n%s",
							stationName,
							stationAddr,
							distance*1000,
							linesStr,
						),
					)
					if _, err = bot.ReplyMessage(event.ReplyToken, postMessage).Do(); err != nil {
						log.Print(err)
					}
				}
			}
		}
	})

	router.Run(":" + port)
}

func getClosestStation(lat float64, lon float64) (station *models.StationByCoordsResponse, err error) {
	client := graphql.NewClient("https://sapi.tinykitten.me/graphql")
	req := graphql.NewRequest(`
    query ($latitude: Float!, $longitude: Float!) {
		stationByCoords(latitude: $latitude, longitude: $longitude) {
		  name
		  address
		  distance
		  lines {
			name
		  }
		}
	  }
	`)
	req.Var("latitude", lat)
	req.Var("longitude", lon)

	ctx := context.Background()

	var result models.StationByCoordsResponse
	if err := client.Run(ctx, req, &result); err != nil {
		log.Fatal(err)
	}

	return &result, nil
}
