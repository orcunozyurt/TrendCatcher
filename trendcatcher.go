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
		truncated := timenow.Truncate(time.Minute)
		query := database.Query{}
		query["tweeted_at"] = truncated

		item, _ := models.GetExpression(query)
		if item != nil {
			count := item.PostCount
			add := *count + 1
			item.PostCount = &add
			item.Update()
			fmt.Println("UPDATED")

		} else {

			expression := &models.Expression{}
			count := 1
			expression.PostCount = &count
			expression.Create()
			fmt.Println("CREATED")

		}

	}
	params := &twitter.StreamFilterParams{
		Track:         []string{"kitten"},
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
