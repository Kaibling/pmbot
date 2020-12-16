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
}

//MultiPlexChannel -
type MultiPlexChannel struct {
	IncomingChannel chan ChannelMessage
	OutgoingChannel chan ChannelMessage
}

func newMultiPlexChannel() MultiPlexChannel {
	returnChannel := MultiPlexChannel{}
	returnChannel.IncomingChannel = make(chan ChannelMessage)
	returnChannel.OutgoingChannel = make(chan ChannelMessage)
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
	Topic   string
	Content interface{}
	Sender  string
}

//NewChannelMessage -
func NewChannelMessage(sender string, topic string) ChannelMessage {
	returnMessage := ChannelMessage{Sender: sender, Topic: topic}
	return returnMessage
}

//Broker -
type Broker struct {
	services      map[string]messageService
	subscriptions map[string][]string
	blocker       chan ChannelMessage
}

//InitBroker -
func InitBroker() *Broker {
	broker := &Broker{}
	broker.services = make(map[string]messageService)
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
	}
	module.SetChannels(pubChannel.invert(), privChannel.invert())
	selfBroker.services[module.GetServiceName()] = messageService
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

	//Broadcast
	for _, value := range selfBroker.subscriptions[message.Topic] {
		selfBroker.sendMessage(value, message)
	}

	if message.Topic == "STATUS" {
		log.Debugf("Status request from %s", message.Sender)

		statusMessage := ""
		statusChannel := make(chan ChannelMessage)
		var wg sync.WaitGroup
		wg.Add(len(selfBroker.services))
		for key := range selfBroker.services {
			go selfBroker.checkStatus(key, statusChannel, &wg)
		}

		for i := 0; i < len(selfBroker.services); i++ {
			select {
			case response := <-statusChannel:
				if response.Content.(string) != "OK" {
					statusMessage += response.Sender + ":" + response.Content.(string)
				}
			}
		}
		wg.Wait()
		if statusMessage == "" {
			statusMessage = "OK"
		}
		log.Debugf("send message to %#v", selfBroker.services[message.Sender].privateChannel.OutgoingChannel)
		selfBroker.services[message.Sender].privateChannel.OutgoingChannel <- ChannelMessage{Topic: "STATUS", Content: statusMessage}
		log.Debugf("Health status send to module")
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
	log.Debugf("Status request to module reddit %v", selfBroker.services[serviceName].publicChannel.IncomingChannel)
	selfBroker.services[serviceName].publicChannel.OutgoingChannel <- ChannelMessage{Topic: "STATUS"}

	timeout := make(chan bool, 1)
	go func() {
		time.Sleep(2 * time.Second)
		timeout <- true
	}()
	overallStatus := "OK"
	select {
	case statusMessage := <-selfBroker.services[serviceName].privateChannel.IncomingChannel:
		log.Debugf("response from %s with %#v", statusMessage.Sender, statusMessage.Content)
	case <-timeout:
		overallStatus = "Timeout from " + serviceName
	}
	statusChannel <- ChannelMessage{Sender: serviceName, Content: overallStatus}
	wg.Done()
}
