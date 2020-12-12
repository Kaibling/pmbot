package main

import (
	"os"
	"os/signal"
	"pmbBot/discord"
	"pmbBot/reddit"
	"syscall"

	log "github.com/sirupsen/logrus"
)

type Config struct {
	Reddit       RedditConfig
	DiscordToken string
}
type RedditConfig struct {
	ClientID string
	Secret   string
	Username string
	Password string
}

var version = "0.1"
var appName = "redditDiscordBot"

func init() {
	log.SetReportCaller(true)
	log.SetLevel(log.DebugLevel)
}

func main() {

	c1 := make(chan string)
	sc := make(chan os.Signal, 1)

	config := Config{
		Reddit: RedditConfig{
			ClientID: os.Getenv("REDDIT_CLIENT_ID"),
			Secret:   os.Getenv("REDDIT_SECRET"),
			Username: os.Getenv("REDDIT_USERNAME"),
			Password: os.Getenv("REDDIT_PASSWORD"),
		},
		DiscordToken: os.Getenv("DISCORD_TOKEN"),
	}

	redditBot := reddit.InitRedditBot(c1,
		config.Reddit.Username,
		config.Reddit.ClientID,
		config.Reddit.Secret,
		config.Reddit.Password,
		version)

	discordBot := discord.InitBot(c1, sc, config.DiscordToken)

	go discordBot.Start()
	go redditBot.Start()

	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	for {
		select {
		case msg1 := <-c1:
			log.Infoln(msg1)
			discordBot.SendNewData(msg1, "786978601891135519")

		case msg1 := <-sc:
			log.Infoln("Stopp die scheiße", msg1)
			discordBot.Stop()
			return
		}
	}

}
