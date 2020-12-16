package discord

import (
	"pmBot/broker"
	"pmBot/configuration"
	"sync"

	"github.com/bwmarrin/discordgo"
	log "github.com/sirupsen/logrus"
)

//DiscBot struct holds data
type DiscBot struct {
	name           string
	bot            *discordgo.Session
	c              <-chan broker.ChannelMessage
	publicChannel  broker.MultiPlexChannel
	privateChannel broker.MultiPlexChannel
	wg             *sync.WaitGroup
}

//SendNewData sends the message to the channel
func (selfDiscBot *DiscBot) SendNewData(data string, channelID string) {
	selfDiscBot.bot.ChannelMessageSend(channelID, data)
	log.Infof("send %s to %s\n", data, channelID)
}

//InitModule configures Discord bot
func InitModule(c <-chan broker.ChannelMessage, token string) *DiscBot {
	log.Infoln("https://discord.com/oauth2/authorize?client_id=194006667195580416&scope=bot")
	dg, err := discordgo.New("Bot " + token)
	if err != nil {
		log.Errorln("Error creating Discord session: ", err)
		return nil
	}
	returBot := &DiscBot{bot: dg, c: c, name: "DISCORD"}
	dg.AddHandler(returBot.messageCreate)
	return returBot
}

//Start Starts the discord bot
func (selfDiscBot *DiscBot) Start(wg *sync.WaitGroup) {
	err := selfDiscBot.bot.Open()
	if err != nil {
		log.Errorln("Error starting server: ", err)
		return
	}
	selfDiscBot.wg = wg

	log.Infoln("discord module is now running...")
	for {
		request := <-selfDiscBot.publicChannel.IncomingChannel
		log.Debugf("request: %#v", request)
		if request.Topic == "REDDIT" {
			selfDiscBot.SendNewData(request.Content.(string), "786978601891135519")
		}
		if request.Topic == "STATUS" {
			selfDiscBot.privateChannel.OutgoingChannel <- broker.ChannelMessage{Topic: "STATUS", Sender: selfDiscBot.name, Content: "OK"}
			log.Debugf("privateChannel: Healthcheck fine ")
		}

	}
}

//Stop stops the discordserver
func (selfDiscBot *DiscBot) Stop() {
	selfDiscBot.bot.Close()
	log.Infoln("Discort bot stopped")
	selfDiscBot.wg.Done()
}
func (selfDiscBot *DiscBot) GetServiceName() string {
	return selfDiscBot.name
}
func (selfDiscBot *DiscBot) SetChannels(pubChannel broker.MultiPlexChannel, privChannel broker.MultiPlexChannel) {
	selfDiscBot.publicChannel = pubChannel
	selfDiscBot.privateChannel = privChannel
}

func (selfDiscBot *DiscBot) messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}

	if m.GuildID == "" {
		log.Infof("Private Message %s %s", m.ChannelID, m.Author.Username)
		if m.Content == "version" {
			s.ChannelMessageSend(m.ChannelID, configuration.Configuration.Variables["Version"])
		}
		if m.Content == "status" {
			log.Debugf("Status request from %s\n sending to %v", m.Author.Username, selfDiscBot.privateChannel.OutgoingChannel)
			statusRequest := broker.NewChannelMessage(selfDiscBot.name, "STATUS")
			selfDiscBot.privateChannel.OutgoingChannel <- statusRequest
			log.Debugf("waiting for broker on %#v", selfDiscBot.privateChannel.IncomingChannel)
			response := <-selfDiscBot.privateChannel.IncomingChannel
			log.Debugf("Response from %#v %#v", selfDiscBot.privateChannel, response)
			s.ChannelMessageSend(m.ChannelID, response.Content.(string))
			log.Debugf("status message sent to %s %s\n", m.ChannelID, m.Author.Username)
		}
		return
	}

	log.Infof("%s %s", m.ChannelID, m.Author.Username)
	if m.ChannelID == "786978601891135519" {
		response := m.Author.Username + ": " + m.Content + "\n"
		s.ChannelMessageSend(m.ChannelID, response)
		log.Infof("send %s to %s\n", response, m.ChannelID)
	}
}
