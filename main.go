package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/line/line-bot-sdk-go/linebot"
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
					nearStation, err := getNearStation(message.Latitude, message.Longitude)
					if err != nil {
						postMessage := linebot.NewTextMessage("ERROR: " + err.Error())
						if _, err = bot.ReplyMessage(event.ReplyToken, postMessage).Do(); err != nil {
							log.Print(err)
						}
					}

					postMessage := linebot.NewTextMessage("最寄り駅は、" + nearStation.StationName + "駅です。")
					if _, err = bot.ReplyMessage(event.ReplyToken, postMessage).Do(); err != nil {
						log.Print(err)
					}
				case *linebot.TextMessage:
					// source := event.Source
					postMessage := linebot.NewTextMessage("OK: " + message.Text)
					if _, err = bot.ReplyMessage(event.ReplyToken, postMessage).Do(); err != nil {
						log.Print(err)
					}
				}
			}
		}
	})

	router.Run(":" + port)
}

// NearStationResponse StationAPI Response
type NearStationResponse struct {
	Addr     string         `json:"addr"`
	CloseYmd sql.NullString `json:"close_ymd"`
	ESort    int            `json:"e_sort"`
	EStatus  int            `json:"e_status"`
	Gap      float64        `json:"gap"`
	Lat      float64        `json:"lat"`
	LineCd   int            `json:"line_cd"`
	Lines    []struct {
		CompanyCd  int            `json:"company_cd"`
		ESort      int            `json:"e_sort"`
		EStatus    int            `json:"e_status"`
		Lat        float64        `json:"lat"`
		LineCd     int            `json:"line_cd"`
		LineColorC string         `json:"line_color_c"`
		LineColorT sql.NullString `json:"line_color_t"`
		LineName   string         `json:"line_name"`
		LineNameH  string         `json:"line_name_h"`
		LineNameK  string         `json:"line_name_k"`
		LineType   sql.NullString `json:"line_type"`
		Lon        float64        `json:"lon"`
		Zoom       int            `json:"zoom"`
	} `json:"lines"`
	Lon          float64        `json:"lon"`
	OpenYmd      sql.NullString `json:"open_ymd"`
	Post         string         `json:"post"`
	PrefCd       int            `json:"pref_cd"`
	StationCd    int            `json:"station_cd"`
	StationGCd   int            `json:"station_g_cd"`
	StationName  string         `json:"station_name"`
	StationNameK sql.NullString `json:"station_name_k"`
	StationNameR sql.NullString `json:"station_name_r"`
}

func getNearStation(lat float64, lon float64) (station *NearStationResponse, err error) {
	client := &http.Client{}
	latStr := strconv.FormatFloat(lat, 'f', 10, 64)
	lonStr := strconv.FormatFloat(lon, 'f', 10, 64)
	resp, err := client.Get("https://sapi.tinykitten.me/v1/station/near?" + latStr + "&lon=" + lonStr)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var result NearStationResponse
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}
