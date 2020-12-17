package broker

import (
	"reflect"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

type messageService struct {
	publicChannel  *MultiPlexChannel
	privateChannel *MultiPlexChannel
	active         bool
}

//MultiPlexChannel -
type MultiPlexChannel struct {
	IncomingChannel chan ChannelMessage
	OutgoingChannel chan ChannelMessage
}

func newMultiPlexChannel() MultiPlexChannel {
	returnChannel := MultiPlexChannel{}
	returnChannel.IncomingChannel = make(chan ChannelMessage, 2)
	returnChannel.OutgoingChannel = make(chan ChannelMessage, 2)
	return returnChannel
}

func (selfChannel *MultiPlexChannel) invert() MultiPlexChannel {
	returnChannel := MultiPlexChannel{}
	returnChannel.OutgoingChannel = selfChannel.IncomingChannel
	returnChannel.IncomingChannel = selfChannel.OutgoingChannel
	return returnChannel
}
func (selfChannel *MultiPlexChannel) send(message ChannelMessage) {
	selfChannel.OutgoingChannel <- message
}

func (selfChannel *MultiPlexChannel) receive() ChannelMessage {
	return <-selfChannel.IncomingChannel
}

//Module -
type Module interface {
	Start(*sync.WaitGroup)
	Stop()
	GetServiceName() string
	SetChannels(publicChannel MultiPlexChannel, privatChannel MultiPlexChannel)
}

//ChannelMessage -
type ChannelMessage struct {
	Topic          string
	Content        interface{}
	Sender         string
	Receiver       string
	OriginalSender string
}

//NewChannelTopicMessage -
func NewChannelTopicMessage(sender string, topic string) ChannelMessage {
	returnMessage := ChannelMessage{Sender: sender, Topic: topic}
	return returnMessage
}

//NewChannelMessage -
func NewChannelMessage(sender, topic, content string) ChannelMessage {
	returnMessage := ChannelMessage{Sender: sender, Topic: topic, Content: content}
	return returnMessage
}

//Broker -
type Broker struct {
	services      map[string]*messageService
	subscriptions map[string][]string
	blocker       chan ChannelMessage
}

//InitBroker -
func InitBroker() *Broker {
	broker := &Broker{}
	broker.services = make(map[string]*messageService)
	broker.subscriptions = make(map[string][]string)
	broker.blocker = make(chan ChannelMessage)
	return broker
}

//AddService -
func (selfBroker *Broker) AddService(module Module) {

	pubChannel := newMultiPlexChannel()
	privChannel := newMultiPlexChannel()

	messageService := messageService{
		publicChannel:  &pubChannel,
		privateChannel: &privChannel,
		active:         true,
	}
	module.SetChannels(pubChannel.invert(), privChannel.invert())
	selfBroker.services[module.GetServiceName()] = &messageService
	log.Debugf("saved module %s in broker \n", module.GetServiceName())
	log.Debugf("pub: %#v priv: %v", pubChannel, privChannel)

}

//SubscribeTopic -
func (selfBroker *Broker) SubscribeTopic(serviceName, topic string) error {
	if _, ok := selfBroker.services[serviceName]; !ok {
		selfBroker.subscriptions[topic] = []string{serviceName}
		return nil
	}
	selfBroker.subscriptions[topic] = append(selfBroker.subscriptions[topic], serviceName)
	return nil
}

func (selfBroker *Broker) sendMessage(serviceName string, message ChannelMessage) {
	log.Debugf("send to %s", serviceName)
	val, ok := selfBroker.services[serviceName]
	if !ok {
		log.Errorf("no service %s saved in broker", serviceName)
		return
	}
	val.publicChannel.send(message)
}

func (selfBroker *Broker) processMessage(message ChannelMessage) {

	if message.Receiver == "" {
		//Broadcast
		if message.OriginalSender == "" {
			for _, value := range selfBroker.subscriptions[message.Topic] {
				if selfBroker.services[value].active {
					selfBroker.sendMessage(value, message)
				}

			}
		} else {
			selfBroker.services[message.OriginalSender].privateChannel.send(message)
		}

	} else {
		selfBroker.sendMessage(message.Receiver, message)
	}

	if message.Topic == "STATUS" {
		log.Debugf("Status request from %s", message.Sender)

		statusMessage := ""
		statusChannel := make(chan ChannelMessage)
		var wg sync.WaitGroup
		goroutineCount := 0
		for key, val := range selfBroker.services {
			if val.active {
				goroutineCount++
			} else {
				statusMessage += key + ":\t inactive\n"
			}
		}

		wg.Add(goroutineCount)
		for key, val := range selfBroker.services {
			if val.active {
				go selfBroker.checkStatus(key, statusChannel, &wg)
			} else {
				log.Debugf("skipping inactive service %s", key)
			}
		}

		for i := 0; i < goroutineCount; i++ {
			select {
			case response := <-statusChannel:
				statusMessage += response.Sender + ":\t" + response.Content.(string) + "\n"
			}
		}
		wg.Wait()
		selfBroker.services[message.Sender].privateChannel.OutgoingChannel <- ChannelMessage{Topic: "STATUS", Content: statusMessage}
		log.Debugf("send message to %#v", selfBroker.services[message.Sender].privateChannel.OutgoingChannel)
	}
}

func recvFromAny(chs []chan ChannelMessage) (val ChannelMessage, from int) {
	set := []reflect.SelectCase{}
	for _, ch := range chs {
		set = append(set, reflect.SelectCase{
			Dir:  reflect.SelectRecv,
			Chan: reflect.ValueOf(ch),
		})
	}
	from, valValue, _ := reflect.Select(set)
	val = valValue.Interface().(ChannelMessage)
	return
}

//Stop -
func (selfBroker *Broker) Stop() {
	selfBroker.blocker <- ChannelMessage{Topic: "SHUTDOWN"}
}

//Start -
func (selfBroker *Broker) Start() {

	//Debug for routing
	log.Debugf("routes %d", len(selfBroker.subscriptions))
	for key, val := range selfBroker.subscriptions {
		log.Debugf("Subscription topic: %s service: %s", key, val)
	}

	channels := []chan ChannelMessage{selfBroker.blocker}
	for _, val := range selfBroker.services {
		channels = append(channels, val.privateChannel.IncomingChannel)
		channels = append(channels, val.publicChannel.IncomingChannel)
	}
	log.Infoln("Broker module is now running...")
	for {
		message, x := recvFromAny(channels)
		log.Debugf("Received %v from ch%v", message, x)
		if message.Topic == "SHUTDOWN" && channels[x] == selfBroker.blocker {
			log.Info("Broker stopped")
			return
		}
		selfBroker.processMessage(message)
	}
}

func (selfBroker *Broker) checkStatus(serviceName string, statusChannel chan ChannelMessage, wg *sync.WaitGroup) {
	log.Debugf("Status request to module %s %v", serviceName, selfBroker.services[serviceName].publicChannel.IncomingChannel)
	selfBroker.services[serviceName].publicChannel.OutgoingChannel <- ChannelMessage{Topic: "STATUS"}

	timeout := make(chan bool, 1)
	go func() {
		time.Sleep(1 * time.Second)
		timeout <- true
	}()
	var overallStatus string
	select {

	case statusMessage := <-selfBroker.services[serviceName].privateChannel.IncomingChannel:
		log.Debugf("response from %s with %#v", statusMessage.Sender, statusMessage.Content)
		overallStatus = statusMessage.Content.(string)
	case <-timeout:
		log.Debugf("Timeout from %s", serviceName)
		//remove service from broker
		overallStatus = "Timeout from " + serviceName
		selfBroker.services[serviceName].active = false

	}

	statusChannel <- ChannelMessage{Sender: serviceName, Content: overallStatus}
	wg.Done()
}
