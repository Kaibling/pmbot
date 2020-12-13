package discord

import (
	"encoding/json"
	"os"

	"github.com/bwmarrin/discordgo"
	log "github.com/sirupsen/logrus"
)

//DiscBot struct holds data
type DiscBot struct {
	bot *discordgo.Session
	c   <-chan string
	sc  <-chan os.Signal
}

//SendNewData sends the message to the channel
func (selfDiscBot *DiscBot) SendNewData(data string, channelID string) {
	selfDiscBot.bot.ChannelMessageSend(channelID, data)
	log.Infof("send %s to %s\n", data, channelID)
}

//Start Starts the discord bot
func (selfDiscBot *DiscBot) Start() {
	err := selfDiscBot.bot.Open()
	if err != nil {
		log.Errorln("Error starting server: ", err)
		return
	}

	// Wait here until CTRL-C or other term signal is received.
	log.Infoln("Bot is now running.  Press CTRL-C to exit.")
	<-selfDiscBot.sc
	selfDiscBot.bot.Close()
}

//Stop stops the discordserver
func (selfDiscBot *DiscBot) Stop() {
	selfDiscBot.bot.Close()
}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}

	if m.GuildID == "" {
		log.Infof("Private Message %s %s", m.ChannelID, m.Author.Username)
		return
	}

	/*
		c, err := s.State.Channel(m.ChannelID)
		if err != nil {
			log.Error(err)
			return
		}
	*/

	/*
		g, err := s.State.Guild(c.GuildID)
		if err != nil {
			log.Error(err)
			return
		}
	*/
	log.Infof("%s %s", m.ChannelID, m.Author.Username)
	if m.ChannelID == "786978601891135519" {
		response := m.Author.Username + ": " + m.Content + "\n"
		s.ChannelMessageSend(m.ChannelID, response)
		log.Infof("send %s to %s\n", response, m.ChannelID)

	}
}

//InitBot configures Discord bot
func InitBot(c <-chan string, sc <-chan os.Signal, token string) *DiscBot {
	log.Infoln("https://discord.com/oauth2/authorize?client_id=194006667195580416&scope=bot")
	dg, err := discordgo.New("Bot " + token)
	if err != nil {
		log.Errorln("Error creating Discord session: ", err)
		return nil
	}
	dg.AddHandler(messageCreate)
	//dg.Identify.Intents = discordgo.MakeIntent(discordgo.IntentsGuildMessages | discordgo.IntentsDirectMessages)
	return &DiscBot{bot: dg, c: c}
}

func pretty(data interface{}) string {
	a, _ := json.MarshalIndent(data, "", " ")
	return string(a)
}
