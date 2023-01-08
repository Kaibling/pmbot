package modules

import (
	"github.com/Kaibling/pmbot/lib/broker"
	"github.com/Kaibling/pmbot/models"

	log "github.com/sirupsen/logrus"
)

type PMModule struct {
	Name   string
	broker *broker.Channel
	topics map[string]chan models.Message
}

func NewPMModule(topics []string, name string) *PMModule {
	t := make(map[string]chan models.Message)
	for _, topic := range topics {
		t[topic] = make(chan models.Message, 2)
	}
	return &PMModule{Name: name, topics: t}
}

func (pm *PMModule) Send(m models.Message) {
	pm.broker.Send(m)
}

func (pm *PMModule) Receive(topic string) models.Message {
	if val, ok := pm.topics[topic]; ok {
		return <-val
	}
	log.Debugf("topic %s not listened ", topic)
	//return <-pm.topics[topic]
	return models.Message{}
}

func (pm *PMModule) Dispatch() {
	for {
		m := pm.broker.Receive()
		log.Debugf("received from dispatcher %#v", m)
		if topicChan, ok := pm.topics[m.Topic]; ok {
			topicChan <- m
			log.Debugf("send to topic channel%#v", m)
		} else {
			log.Debugf("default hilfer????")
		}
	}
}

func (pm *PMModule) SetChannel(c *broker.Channel) {
	pm.broker = c
	pm.broker.Active = true
}
func (pm *PMModule) Active() bool {
	return pm.broker.Active
}
func (pm *PMModule) GetTopics() (topics []string) {
	for t, _ := range pm.topics {
		topics = append(topics, t)
	}
	return
}
