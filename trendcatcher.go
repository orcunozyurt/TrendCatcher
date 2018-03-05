package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"
	"github.com/joho/godotenv"
	"github.com/trendcatcher/database"
	"github.com/trendcatcher/models"
	"github.com/trendcatcher/utils"
	"github.com/tuvistavie/structomap"
)

func init() {
	database.Connect()
	database.EnsureIndexes()
	// Use snake case in all serializers
	structomap.SetDefaultCase(structomap.SnakeCase)
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file", err)
	}
}

func initTwitterConnection() *twitter.Client {
	fmt.Println("Initializing Twitter Connection...")

	consumerKey := utils.GetEnvOrDefault("TWITTER_CONSUMER_KEY", "")
	consumerSecret := utils.GetEnvOrDefault("TWITTER_CONSUMER_SECRET", "")

	accessToken := utils.GetEnvOrDefault("TWITTER_ACCESS_TOKEN", "")
	accessSecret := utils.GetEnvOrDefault("TWITTER_ACCESS_SECRET", "")

	config := oauth1.NewConfig(consumerKey, consumerSecret)
	token := oauth1.NewToken(accessToken, accessSecret)
	httpClient := config.Client(oauth1.NoContext, token)

	return twitter.NewClient(httpClient)

}

func main() {

	client := initTwitterConnection()
	demux := twitter.NewSwitchDemux()
	demux.Tweet = func(tweet *twitter.Tweet) {

		timenow := time.Now()
		truncatedMinute := timenow.Truncate(time.Minute)
		truncatedHour := timenow.Truncate(time.Hour)
		truncatedDay := time.Date(timenow.Year(), timenow.Month(), timenow.Day(), 0, 0, 0, 0, timenow.Location())

		querymin := database.Query{}
		querymin["tweeted_at"] = truncatedMinute

		queryhour := database.Query{}
		queryhour["tweeted_at"] = truncatedHour

		queryday := database.Query{}
		queryday["tweeted_at"] = truncatedDay

		itemMinute, _ := models.GetExpression(querymin, 0)
		itemHour, _ := models.GetExpression(queryhour, 1)
		itemDay, _ := models.GetExpression(queryday, 2)
		if itemMinute != nil {
			count := itemMinute.PostCount
			add := *count + 1
			itemMinute.PostCount = &add
			itemMinute.Update(0)
			fmt.Println("UPDATED", "PERMINUTE")

		} else {

			expression := &models.Expression{}
			count := 1
			expression.PostCount = &count
			expression.Create(0)
			fmt.Println("CREATED", "PERMINUTE")

		}

		if itemHour != nil {
			count := itemHour.PostCount
			add := *count + 1
			itemHour.PostCount = &add
			itemHour.Update(1)
			fmt.Println("UPDATED", "HOURLY")

		} else {

			expression := &models.Expression{}
			count := 1
			expression.PostCount = &count
			expression.Create(1)
			fmt.Println("CREATED", "HOURLY")

		}

		if itemDay != nil {
			count := itemDay.PostCount
			add := *count + 1
			itemDay.PostCount = &add
			itemDay.Update(2)
			fmt.Println("UPDATED", "DAILY")

		} else {

			expression := &models.Expression{}
			count := 1
			expression.PostCount = &count
			expression.Create(2)
			fmt.Println("CREATED", "DAILY")

		}

	}
	params := &twitter.StreamFilterParams{
		Track:         []string{"bitcoin"},
		StallWarnings: twitter.Bool(true),
	}
	stream, err := client.Streams.Filter(params)
	if err != nil {
		fmt.Println(err)
	}
	go demux.HandleChan(stream.Messages)

	// Wait for SIGINT and SIGTERM (HIT CTRL-C)
	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	log.Println(<-ch)

	stream.Stop()

}
