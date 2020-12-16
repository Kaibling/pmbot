package main

import (
	"os"
	"os/signal"
	"pmbBot/broker"
	"pmbBot/configuration"
	"pmbBot/discord"
	"pmbBot/reddit"
	"sync"
	"syscall"

	log "github.com/sirupsen/logrus"
)

var version string
var appName = "redditDiscordBot"
var buildTime string

func init() {
	log.SetReportCaller(true)
	//log.SetFormatter(&log.JSONFormatter{})
	log.SetLevel(log.DebugLevel)

	//CONFIG
	config := configuration.Config{
		Reddit: configuration.RedditConfig{
			ClientID: os.Getenv("REDDIT_CLIENT_ID"),
			Secret:   os.Getenv("REDDIT_SECRET"),
			Username: os.Getenv("REDDIT_USERNAME"),
			Password: os.Getenv("REDDIT_PASSWORD"),
		},
		DiscordToken: os.Getenv("DISCORD_TOKEN"),
		Variables:    make(map[string]string),
	}
	config.Variables["Version"] = version
	config.Variables["subreddit"] = "AskReddit"
	configuration.Apply(config)
	log.Infof("Version: %s", version)
	log.Infof("Buildtime: %s", buildTime)
}

func main() {
	c1 := make(chan broker.ChannelMessage)
	sc := make(chan os.Signal, 1)

	//INIT MODULES
	modules := []broker.Module{}

	//reddit

	modules = append(modules, reddit.InitModule(c1,
		configuration.Configuration.Reddit.Username,
		configuration.Configuration.Reddit.ClientID,
		configuration.Configuration.Reddit.Secret,
		configuration.Configuration.Reddit.Password,
		configuration.Configuration.Variables["subreddit"]))

	//discord
	modules = append(modules, discord.InitModule(c1, configuration.Configuration.DiscordToken))

	brokerInstance := broker.InitBroker()
	brokerInstance.SubscribeTopic("DISCORD", "REDDIT")

	//START MODULES
	var wg sync.WaitGroup
	wg.Add(len(modules))
	for _, module := range modules {
		brokerInstance.AddService(module)
		go module.Start(&wg)
	}
	go brokerInstance.Start()
	log.Infoln("All modules loaded. Press CTRL-C to exit.")
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	for {
		msg1 := <-sc
		//brokerInstance.SendMessage("DISCORD", broker.ChannelMessage{Topic: "", Content: "HEY hO"})
		for _, module := range modules {
			go module.Stop()
		}
		wg.Wait()
		log.Infoln("Stopp die scheiße", msg1)
		return
	}
}
