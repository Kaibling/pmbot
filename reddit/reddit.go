package reddit

import (
	"github.com/turnage/graw"
	"github.com/turnage/graw/reddit"

	log "github.com/sirupsen/logrus"
)

//GrabBot sds
type GrabBot struct {
	bot reddit.Bot
	c   chan<- string
}

//Post -
func (r *GrabBot) Post(post *reddit.Post) error {
	r.c <- post.Title + "\n" + post.URL
	return nil
}

//Start -
func (r *GrabBot) Start() {
	cfg := graw.Config{Subreddits: []string{"FreeGameFindings"}}
	_, wait, err := graw.Run(r, r.bot, cfg)
	if err != nil {
		log.Errorln("Failed to start graw run: ", err)
	} else {
		log.Infoln("Redditbot started")
		wait()
	}
}

//InitRedditBot -
func InitRedditBot(c chan<- string, username string, clientID string, secret string, password string, version string) *GrabBot {

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
