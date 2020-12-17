package discord

import (
	"pmbot/broker"
	"pmbot/configuration"
	"strings"
	"sync"

	"github.com/bwmarrin/discordgo"
	log "github.com/sirupsen/logrus"
)

//DiscBot struct holds data
type DiscBot struct {
	name           string
	bot            *discordgo.Session
	publicChannel  *broker.MultiPlexChannel
	privateChannel *broker.MultiPlexChannel
	wg             *sync.WaitGroup
}

//SendNewData sends the message to the channel
func (selfDiscBot *DiscBot) sendNewData(data string, channelID string) {
	selfDiscBot.bot.ChannelMessageSend(channelID, data)
	log.Infof("send Data to  %s\n", channelID)
	log.Debugf("send %#v\n", data)
}

//InitModule configures Discord bot
func InitModule(token string) *DiscBot {
	log.Infoln("https://discord.com/oauth2/authorize?client_id=194006667195580416&scope=bot")
	dg, err := discordgo.New("Bot " + token)
	if err != nil {
		log.Errorln("Error creating Discord session: ", err)
		return nil
	}
	returBot := &DiscBot{bot: dg, name: "DISCORD"}
	dg.AddHandler(returBot.messageCreate)
	return returBot
}

//Start Starts the discord bot
func (selfDiscBot *DiscBot) Start(wg *sync.WaitGroup) {
	selfDiscBot.wg = wg
	err := selfDiscBot.bot.Open()
	if err != nil {
		log.Errorln("Error starting server: ", err)
		return
	}
	if selfDiscBot.publicChannel == nil || selfDiscBot.privateChannel == nil {
		log.Errorf("Channels not initilized %#v %#v", selfDiscBot.publicChannel, selfDiscBot.privateChannel)
		return
	}

	log.Infoln("discord module is now running...")
	for {
		request := <-selfDiscBot.publicChannel.IncomingChannel
		log.Debugf("request: %#v", request)
		if request.Topic == "REDDIT" {
			selfDiscBot.sendNewData(request.Content.(string), "786978601891135519")
		}
		if request.Topic == "STATUS" {
			selfDiscBot.privateChannel.OutgoingChannel <- broker.ChannelMessage{Topic: "STATUS", Sender: selfDiscBot.name, Content: "OK"}
			log.Debugf("Healthcheck fine")
		}

	}
}

//Stop stops the discordserver
func (selfDiscBot *DiscBot) Stop() {
	if selfDiscBot.bot == nil {
		log.Errorf("bot cannot be closed and not be started")
	} else {
		selfDiscBot.bot.Close()
	}
	if selfDiscBot.wg == nil {
		log.Errorf("Waitgroup cannot be done, if not started")
	} else {
		selfDiscBot.wg.Done()
	}
	log.Infoln("Discord bot stopped")
}
func (selfDiscBot *DiscBot) GetServiceName() string {
	return selfDiscBot.name
}
func (selfDiscBot *DiscBot) SetChannels(pubChannel broker.MultiPlexChannel, privChannel broker.MultiPlexChannel) {
	selfDiscBot.publicChannel = &pubChannel
	selfDiscBot.privateChannel = &privChannel
}

func (selfDiscBot *DiscBot) messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}

	if m.GuildID == "" {
		log.Infof("Private Message %s %s", m.ChannelID, m.Author.Username)
		command := strings.Split(m.Content, " ")
		if m.Content == "version" {
			selfDiscBot.sendNewData(configuration.Configuration.Variables["Version"], m.ChannelID)
		}
		if m.Content == "status" {
			log.Infof("Status request from %s", m.Author.Username)
			statusRequest := broker.NewChannelTopicMessage(selfDiscBot.name, "STATUS")
			selfDiscBot.privateChannel.OutgoingChannel <- statusRequest

			log.Debugf("waiting for broker on %#v", selfDiscBot.privateChannel.IncomingChannel)
			response := <-selfDiscBot.privateChannel.IncomingChannel

			log.Debugf("Response from %#v %#v", selfDiscBot.privateChannel, response)
			selfDiscBot.sendNewData(response.Content.(string), m.ChannelID)
		}
		if command[0] == "harvest" {
			statusRequest := broker.ChannelMessage{Sender: selfDiscBot.name, Topic: "REDDIT_HARVEST", Content: command[1], Receiver: "REDDIT", OriginalSender: "DISCORD"}
			selfDiscBot.privateChannel.OutgoingChannel <- statusRequest

			log.Debugf("waiting for broker on %#v", selfDiscBot.privateChannel.IncomingChannel)
			response := <-selfDiscBot.privateChannel.IncomingChannel

			selfDiscBot.sendNewData(response.Content.(string), m.ChannelID)
		}
		return
	}

	log.Infof("%s %s", m.ChannelID, m.Author.Username)
	if m.ChannelID == "786978601891135519" {
		response := m.Author.Username + ": " + m.Content + "\n"
		selfDiscBot.sendNewData(response, "786978601891135519")
	}
}
