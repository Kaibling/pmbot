package main

import (
	"os"
	"os/signal"
	"pmbBot/discord"
	"pmbBot/reddit"
	"pmbBot/utils"
	"sync"
	"syscall"

	log "github.com/sirupsen/logrus"
)

type config struct {
	Reddit       redditConfig
	DiscordToken string
}
type redditConfig struct {
	ClientID string
	Secret   string
	Username string
	Password string
}

type module interface {
	Start(*sync.WaitGroup)
	Stop()
}

var version = "0.1"
var appName = "redditDiscordBot"

func init() {
	log.SetReportCaller(true)
	log.SetFormatter(&log.JSONFormatter{})
	log.SetLevel(log.DebugLevel)
}

func main() {
	c1 := make(chan utils.ChannelMessage)
	sc := make(chan os.Signal, 1)

	//CONFIG
	config := config{
		Reddit: redditConfig{
			ClientID: os.Getenv("REDDIT_CLIENT_ID"),
			Secret:   os.Getenv("REDDIT_SECRET"),
			Username: os.Getenv("REDDIT_USERNAME"),
			Password: os.Getenv("REDDIT_PASSWORD"),
		},
		DiscordToken: os.Getenv("DISCORD_TOKEN"),
	}

	//INIT MODULES
	modules := []module{}

	//reddit
	modules = append(modules, reddit.InitModule(c1,
		config.Reddit.Username,
		config.Reddit.ClientID,
		config.Reddit.Secret,
		config.Reddit.Password,
		version))

	//discord
	modules = append(modules, discord.InitModule(c1, config.DiscordToken))

	//START MODULES
	var wg sync.WaitGroup
	wg.Add(len(modules))
	for _, module := range modules {
		go module.Start(&wg)
	}

	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	for {
		msg1 := <-sc
		for _, module := range modules {
			go module.Stop()
		}
		wg.Wait()
		log.Infoln("Stopp die scheiße", msg1)
		return
	}
}
