package discord

import (
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
	c, err := s.State.Channel(m.ChannelID)
	if err != nil {
		return
	}

	g, err := s.State.Guild(c.GuildID)
	if err != nil {
		return
	}

	if m.ChannelID == "786978601891135519" {
		if m.Author.ID == s.State.User.ID {
			return
		}
		response := m.Author.Username + ": " + m.Content + "\n" + g.Name + ":" + c.Name
		s.ChannelMessageSend(m.ChannelID, response)
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
	return &DiscBot{bot: dg, c: c}
}
