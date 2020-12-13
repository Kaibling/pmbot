package reddit

import (
	"pmbBot/utils"
	"sync"

	log "github.com/sirupsen/logrus"
	"github.com/turnage/graw"
	"github.com/turnage/graw/reddit"
)

//GrabBot sds
type GrabBot struct {
	bot    reddit.Bot
	c      chan<- utils.ChannelMessage
	stopFn func()
	wg     *sync.WaitGroup
}

//Post -
func (r *GrabBot) Post(post *reddit.Post) error {
	r.c <- utils.ChannelMessage{Topic: "REDDIT_FREE_GAME", Content: post.Title + "\n" + post.URL}
	return nil
}

//Stop -
func (r *GrabBot) Stop() {
	r.stopFn()
	log.Info("redditbot stopped")
	r.wg.Done()
}

//Start -
func (r *GrabBot) Start() {
	cfg := graw.Config{Subreddits: []string{"FreeGameFindings"}}
	_, wait, err := graw.Run(r, r.bot, cfg)
	if err != nil {
		log.Errorln("Failed to start graw run: ", err)
		return
	}
	r.stopFn = stop
	r.wg = wg
	log.Infoln("Redditbot started")
	wait()
}

//InitRedditBot -
func InitModule(c chan<- utils.ChannelMessage, username string, clientID string, secret string, password string, version string) *GrabBot {
	agentString := "grab_data:redditdisc:" + version + " by " + username
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
	return &GrabBot{bot: bot, c: c}
}
