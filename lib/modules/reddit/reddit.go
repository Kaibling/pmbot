package reddit

import (
	"sync"

	"github.com/Kaibling/pmbot/lib/broker"
	"github.com/Kaibling/pmbot/lib/config"
	"github.com/Kaibling/pmbot/lib/modules"
	"github.com/Kaibling/pmbot/models"
	log "github.com/sirupsen/logrus"
	"github.com/turnage/graw"
	"github.com/turnage/graw/reddit"
)

type RedditBot struct {
	pmModule     *modules.PMModule
	SubReddits   []string
	bot          reddit.Bot
	stopFn       func()
	wg           *sync.WaitGroup
	TopicHandler map[string]func(*RedditBot, models.Message)
}

func InitModule(username string, clientID string, secret string, password string, subReddits []string) *RedditBot {
	agentString := "grab_data:redditdisc:" + config.Configuration.Version + " by " + username
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
	funcMapping := map[string]func(*RedditBot, models.Message){
		// "REDDIT_HARVEST": func(r *RedditBot, m models.Message) {
		// 	harvest := r.harvestSubreddit(m.ContentStr())
		// 	log.Println(harvest.Posts)
		// 	log.Println(len(harvest.Posts))
		// 	result := ""
		// 	for _, post := range harvest.Posts[:3] {
		// 		result += fmt.Sprintf(`[%d] %s posted "%s"\n`, post.CreatedUTC, post.Author, post.Title)
		// 		//log.Printf(`[%d] %s posted "%s"\n`, post.CreatedUTC, post.Author, post.Title)
		// 	}
		// 	r.pmModule.Send(models.Message{Topic: "REDDIT_STATUS_RESPONSE", Content: "OK"})
		// 	r.publicChannel.OutgoingChannel <- broker.ChannelMessage{Topic: "REDDIT", Sender: r.name, Content: result}
		// },
		"REDDIT_STATUS_REQUEST": func(r *RedditBot, m models.Message) {
			r.pmModule.Send(models.Message{Topic: "REDDIT_STATUS_RESPONSE", Content: "OK"})
		},
	}
	topicNames := []string{}
	for topicName := range funcMapping {
		topicNames = append(topicNames, topicName)
	}
	topicNames = append(topicNames, "REDDIT_STATUS_RESPONSE")

	return &RedditBot{bot: bot, TopicHandler: funcMapping, pmModule: modules.NewPMModule(topicNames, "REDDIT"), SubReddits: subReddits}
}

// Start -
func (r *RedditBot) Start(wg *sync.WaitGroup) {
	r.wg = wg
	log.Debugf("listining to %#v", r.SubReddits)
	cfg := graw.Config{Subreddits: r.SubReddits}
	//cfg := graw.Config{}
	stop, wait, err := graw.Run(r, r.bot, cfg)
	if err != nil {
		log.Errorln("Failed to start graw run: ", err)
		r.Stop()
		return
	}
	r.stopFn = stop

	log.Infoln("reddit module is now running...")
	go wait()

	for topicName, f := range r.TopicHandler {
		go func(t string, f func(*RedditBot, models.Message)) {
			for {
				log.Debugf("waiting for %s", t)
				m := r.pmModule.Receive(t)
				log.Debugf("got %#v", m)
				f(r, m)
			}

		}(topicName, f)
	}

}

// Stop -
func (r *RedditBot) Stop() {
	if r.stopFn == nil {
		log.Errorf("bot cannot be closed and not be started")
	} else {
		r.stopFn()
	}
	if r.wg == nil {
		log.Errorf("Waitgroup cannot be done, if not started")
	} else {
		r.wg.Done()
	}
	log.Info("redditbot stopped")
}

func (r *RedditBot) GetServiceName() string {
	return r.pmModule.Name
}

func (r *RedditBot) SetChannel(c *broker.Channel) {
	r.pmModule.SetChannel(c)
}

func (r *RedditBot) GetTopics() []string {
	return r.pmModule.GetTopics()
}

// func (r *RedditBot) harvestSubreddit(subRedditName string) *reddit.Harvest {
// 	harvest, err := r.bot.Listing(subRedditName, "")
// 	if err != nil {
// 		log.Println("Failed to fetch "+subRedditName, err)
// 		return nil
// 	}
// 	return &harvest
// }

func (r *RedditBot) Post(post *reddit.Post) error {
	log.Debugf("new post %s", post.Title)
	r.pmModule.Send(models.Message{Topic: "REDDIT_NEW_POST", Content: post.Title + "\n" + post.URL})
	// if post.Subreddit == r.subReddit {
	// 	r.publicChannel.OutgoingChannel <- broker.ChannelMessage{Topic: "REDDIT", Content: post.Title + "\n" + post.URL}
	// }
	return nil
}
