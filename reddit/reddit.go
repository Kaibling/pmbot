package reddit

import (
	"pmbot/broker"
	"pmbot/configuration"
	"sync"

	log "github.com/sirupsen/logrus"
	"github.com/turnage/graw"
	"github.com/turnage/graw/reddit"
)

//GrabBot sds
type GrabBot struct {
	bot            reddit.Bot
	stopFn         func()
	wg             *sync.WaitGroup
	subReddit      string
	name           string
	publicChannel  broker.MultiPlexChannel
	privateChannel broker.MultiPlexChannel
}

//Post -
func (r *GrabBot) Post(post *reddit.Post) error {
	if post.Subreddit == r.subReddit {
		r.publicChannel.OutgoingChannel <- broker.ChannelMessage{Topic: "REDDIT", Content: post.Title + "\n" + post.URL}
	}
	return nil
}

//InitModule -
func InitModule(username string, clientID string, secret string, password string, subReddit string) *GrabBot {
	agentString := "grab_data:redditdisc:" + configuration.Configuration.Variables["Version"] + " by " + username
	botcfg := reddit.BotConfig{
		Agent: agentString,
		App: reddit.App{
			ID:       clientID,
			Secret:   secret,
			Username: username,
			Password: password,
		},
	}
	bot, _ := reddit.NewBot(botcfg)
	return &GrabBot{bot: bot, subReddit: subReddit, name: "REDDIT"}
}

//Start -
func (r *GrabBot) Start(wg *sync.WaitGroup) {
	cfg := graw.Config{Subreddits: []string{r.subReddit}}
	stop, wait, err := graw.Run(r, r.bot, cfg)
	if err != nil {
		log.Errorln("Failed to start graw run: ", err)
		return
	}
	r.stopFn = stop
	r.wg = wg
	log.Infoln("reddit module is now running...")
	go wait()

	for {
		var message broker.ChannelMessage
		select {
		case message = <-r.publicChannel.IncomingChannel:
			log.Debugf("received %s", message.Topic)
		case message = <-r.privateChannel.IncomingChannel:
			log.Debugf("privateChannel: received %s", message.Topic)
		}
		if message.Topic == "STATUS" {
			r.privateChannel.OutgoingChannel <- broker.ChannelMessage{Topic: "STATUS", Sender: r.name, Content: "OK"}
			log.Debugf("privateChannel: Healthcheck fine ")
		}
	}

}

//Stop -
func (r *GrabBot) Stop() {
	r.stopFn()
	log.Info("redditbot stopped")
	r.wg.Done()
}

func (r *GrabBot) GetServiceName() string {
	return r.name
}

func (r *GrabBot) SetChannels(pubChannel broker.MultiPlexChannel, privChannel broker.MultiPlexChannel) {
	r.publicChannel = pubChannel
	r.privateChannel = privChannel
}
