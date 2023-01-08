package discord

import (
	"strings"
	"sync"

	"github.com/Kaibling/pmbot/lib/broker"
	"github.com/Kaibling/pmbot/lib/config"
	"github.com/Kaibling/pmbot/lib/modules"
	"github.com/Kaibling/pmbot/models"
	log "github.com/sirupsen/logrus"

	"github.com/bwmarrin/discordgo"
)

type DiscordBot struct {
	pmModule     *modules.PMModule
	session      *discordgo.Session
	wg           *sync.WaitGroup
	TopicHandler map[string]func(*DiscordBot, models.Message)
}

func (d *DiscordBot) send(data, channelID string) {
	d.session.ChannelMessageSend(channelID, data)
	log.Infof("send Data to  %s\n", channelID)
	log.Debugf("send %#v\n", data)
}

// InitModule configures Discord bot
func InitModule(token string) *DiscordBot {
	session, err := discordgo.New("Bot " + token)
	if err != nil {
		log.Errorln("Error creating Discord session: ", err)
		return nil
	}

	funcMapping := map[string]func(*DiscordBot, models.Message){
		"REDDIT": func(d *DiscordBot, m models.Message) {
			d.send(m.ContentStr(), "786978601891135519")
		},
		"DISCORD_STATUS_REQUEST": func(d *DiscordBot, m models.Message) {
			d.pmModule.Send(models.Message{Topic: "DISCORD_STATUS_RESPONSE", Content: "OK"})
		},
		"REDDIT_NEW_POST": func(d *DiscordBot, m models.Message) {
			d.send(m.ContentStr(), "786978601891135519")
		},
	}

	topicNames := []string{}
	for topicName := range funcMapping {
		topicNames = append(topicNames, topicName)
	}
	topicNames = append(topicNames, "DISCORD_STATUS_RESPONSE")
	topicNames = append(topicNames, "REDDIT_NEW_POST")

	b := &DiscordBot{session: session, TopicHandler: funcMapping, pmModule: modules.NewPMModule(topicNames, "DISCORD")}
	session.AddHandler(b.messageCreate)
	return b
}

func (d *DiscordBot) GetServiceName() string {
	return d.pmModule.Name
}
func (d *DiscordBot) GetTopics() []string {
	return d.pmModule.GetTopics()
}
func (d *DiscordBot) SetChannel(c *broker.Channel) {
	d.pmModule.SetChannel(c)
}

// Start Starts the discord bot
func (d *DiscordBot) Start(wg *sync.WaitGroup) {
	d.wg = wg
	err := d.session.Open()
	if err != nil {
		log.Errorln("Error starting server: ", err)
		return
	}
	if !d.pmModule.Active() {
		log.Errorf("Discord channels not initilized")
		return
	}
	go d.pmModule.Dispatch()
	log.Infoln("discord module is now running...")
	for topicName, f := range d.TopicHandler {
		go func(t string, f func(*DiscordBot, models.Message)) {
			for {
				log.Debugf("waiting for %s", t)
				m := d.pmModule.Receive(t)
				log.Debugf("got %#v", m)
				f(d, m)
			}

		}(topicName, f)
	}
}

// Stop stops the discordserver
func (d *DiscordBot) Stop() {
	if d.session == nil {
		log.Errorf("bot cannot be closed and not be started")
	} else {
		d.session.Close()
	}
	if d.wg == nil {
		log.Errorf("Waitgroup cannot be done, if not started")
	} else {
		d.wg.Done()
	}
	log.Infoln("Discord bot stopped")
}

func (d *DiscordBot) messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	// prevent self messaging
	if m.Author.ID == s.State.User.ID {
		return
	}
	// Privat message
	if m.GuildID == "" {
		sender := m.Author.Username
		chanID := m.ChannelID
		log.Infof("Private Message from %s in channel %s", sender, chanID)
		command := strings.Split(m.Content, " ")
		if len(command) < 1 {
			log.Error("empty command")
			return
		}
		switch strings.ToLower(command[0]) {
		case "version":
			var response string
			if ver, ok := config.Configuration.Variables["Version"]; ok {
				response = ver
			} else {
				response = "no version available"
			}
			d.send(response, chanID)
		case "status":
			log.Infof("Status request from %s", sender)
			d.pmModule.Send(models.Message{Topic: "DISCORD_STATUS_REQUEST"})

			log.Debugf("waiting for broker")
			response := d.pmModule.Receive("DISCORD_STATUS_RESPONSE")
			log.Debugf("Response from %#v", response)
			d.send(response.Content.(string), chanID)
		}

		// if command[0] == "harvest" {
		// 	statusRequest := broker.ChannelMessage{Sender: selfDiscBot.name, Topic: "REDDIT_HARVEST", Content: command[1], Receiver: "REDDIT", OriginalSender: "DISCORD"}
		// 	selfDiscBot.privateChannel.OutgoingChannel <- statusRequest

		// 	log.Debugf("waiting for broker on %#v", selfDiscBot.privateChannel.IncomingChannel)
		// 	response := <-selfDiscBot.privateChannel.IncomingChannel

		// 	selfDiscBot.sendNewData(response.Content.(string), m.ChannelID)
		// }
		return
	}

	// log.Infof("%s %s", m.ChannelID, m.Author.Username)
	// if m.ChannelID == "786978601891135519" {
	// 	response := m.Author.Username + ": " + m.Content + "\n"
	// 	selfDiscBot.sendNewData(response, "786978601891135519")
	// }
}
