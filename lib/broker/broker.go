package broker

import (
	"sync"

	log "github.com/sirupsen/logrus"

	"github.com/Kaibling/pmbot/models"
)

type Module interface {
	Start(*sync.WaitGroup)
	Stop()
	GetServiceName() string
	GetTopics() []string
	SetChannel(*Channel)
}

type Channel struct {
	in     chan models.Message
	out    chan models.Message
	Active bool
}

func newChannel() (c *Channel) {
	c = &Channel{}
	c.in = make(chan models.Message, 2)
	c.out = make(chan models.Message, 2)
	c.Active = true
	return
}

func (c *Channel) invert() (nc *Channel) {
	nc = &Channel{}
	nc.out = c.in
	nc.in = c.out
	return
}
func (c *Channel) Send(m models.Message) {
	c.out <- m
	log.Debugf("send %#v", m)
}

func (c *Channel) Receive() models.Message {
	return <-c.in
}

type Broker struct {
	modules       map[string]*Channel // module -> channel
	subscriptions map[string][]string // topics -> [] modules
	blocker       chan models.Message
}

func NewBroker() (b *Broker) {
	b = &Broker{}
	b.subscriptions = make(map[string][]string)
	b.modules = make(map[string]*Channel)
	b.blocker = make(chan models.Message)
	return
}

func (b *Broker) AddModule(module Module) {
	modChan := newChannel()
	module.SetChannel(modChan.invert())
	b.modules[module.GetServiceName()] = modChan
	for _, topic := range module.GetTopics() {
		b.subscriptions[topic] = append(b.subscriptions[topic], module.GetServiceName())
	}
	log.Debugf("added module %s in broker \n", module.GetServiceName())
}

func (b *Broker) Stop() {
	b.blocker <- models.Message{Topic: "SHUTDOWN"}
}

func (b *Broker) Start() {

	//Debug for routing
	log.Debugf("SUbsciptions %d", len(b.subscriptions))
	for key, val := range b.subscriptions {
		log.Debugf("Subscription topic: %s service: %s", key, val)
	}

	log.Infoln("Broker module is now running...")
	for moduleName, moduleChan := range b.modules {
		go func(moduleName string, moduleChan *Channel) {
			for {
				m := moduleChan.Receive()
				log.Debugf("Received %v from ch%v", m, moduleName)
				b.processMessage(m)
			}
		}(moduleName, moduleChan)
		//message, x := recv(channels)

	}
	<-b.blocker
	// if m.Topic == "SHUTDOWN" {
	// 	log.Info("Broker stopped")
	// 	return
	// }
}

func (b *Broker) processMessage(m models.Message) {
	if val, ok := b.subscriptions[m.Topic]; ok {
		for _, value := range val {
			log.Debugf("and send to %s", value)
			if b.modules[value].Active {
				log.Debugf("send again")
				b.sendMessage(value, m)
			}
		}
	} else {
		log.Debugf("broker has topic %s not subscribed", m.Topic)
	}

	// if message.Topic == "STATUS" {
	// 	log.Debugf("Status request from %s", message.Sender)

	// 	statusMessage := ""
	// 	statusChannel := make(chan ChannelMessage)
	// 	var wg sync.WaitGroup
	// 	goroutineCount := 0
	// 	for key, val := range selfBroker.services {
	// 		if val.active {
	// 			goroutineCount++
	// 		} else {
	// 			statusMessage += key + ":\t inactive\n"
	// 		}
	// 	}

	// 	wg.Add(goroutineCount)
	// 	for key, val := range selfBroker.services {
	// 		if val.active {
	// 			go selfBroker.checkStatus(key, statusChannel, &wg)
	// 		} else {
	// 			log.Debugf("skipping inactive service %s", key)
	// 		}
	// 	}

	// 	for i := 0; i < goroutineCount; i++ {
	// 		select {
	// 		case response := <-statusChannel:
	// 			statusMessage += response.Sender + ":\t" + response.Content.(string) + "\n"
	// 		}
	// 	}
	// 	wg.Wait()
	// 	selfBroker.services[message.Sender].privateChannel.OutgoingChannel <- ChannelMessage{Topic: "STATUS", Content: statusMessage}
	// 	log.Debugf("send message to %#v", selfBroker.services[message.Sender].privateChannel.OutgoingChannel)
	// }
}

func (selfBroker *Broker) sendMessage(moduleName string, m models.Message) {
	log.Debugf("send to %s", moduleName)
	if val, ok := selfBroker.modules[moduleName]; ok {
		val.Send(m)
	} else {
		log.Errorf("no module %s saved in broker", moduleName)
	}
}
