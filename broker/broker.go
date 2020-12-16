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
	//channel chan ChannelMessage
	Sender string
	//MessageID string
	//status    string
}

func NewChannelMessage(sender string, topic string) ChannelMessage {
	returnMessage := ChannelMessage{Sender: sender, Topic: topic}
	//uuidWithHyphen := uuid.New()
	//returnMessage.MessageID = strings.Replace(uuidWithHyphen.String(), "-", "", -1)
	return returnMessage
}

type Broker struct {
	services      map[string]messageService
	subscriptions map[string][]string
}

func InitBroker() *Broker {
	broker := &Broker{}
	broker.services = make(map[string]messageService)
	broker.subscriptions = make(map[string][]string)
	return broker
}
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

func (selfBroker *Broker) SubscribeTopic(serviceName, topic string) {
	selfBroker.subscriptions[topic] = append(selfBroker.subscriptions[topic], serviceName)
}

func (selfBroker *Broker) SendMessage(serviceName string, message ChannelMessage) {
	log.Debugf("send to %s", serviceName)
	selfBroker.services[serviceName].publicChannel.send(message)
	log.Debugf("sent")
}

func (selfBroker *Broker) processMessage(message ChannelMessage) {

	//Broadcast
	for _, value := range selfBroker.subscriptions[message.Topic] {
		selfBroker.SendMessage(value, message)
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

func (selfBroker *Broker) Start() {
	channels := []chan ChannelMessage{}
	for _, val := range selfBroker.services {
		channels = append(channels, val.privateChannel.IncomingChannel)
		channels = append(channels, val.publicChannel.IncomingChannel)
	}

	for {
		message, x := recvFromAny(channels)
		log.Debugf("Received %v from ch%v", message, x)
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
