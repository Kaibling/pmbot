package scheduler

import (
	"pmbot/broker"
	"sync"

	log "github.com/sirupsen/logrus"
)

//DiscBot struct holds data
type Scheduler struct {
	name           string
	publicChannel  *broker.MultiPlexChannel
	privateChannel *broker.MultiPlexChannel
	wg             *sync.WaitGroup
}

func (scheduler *Scheduler) Start(wg *sync.WaitGroup) {
	scheduler.wg = wg
	if scheduler.publicChannel == nil || scheduler.privateChannel == nil {
		log.Errorf("Channels not initilized %#v %#v", scheduler.publicChannel, scheduler.privateChannel)
		return
	}

	log.Infof("%s module is now running...", scheduler.name)
	for {
		request := <-scheduler.publicChannel.IncomingChannel
		log.Debugf("request: %#v", request)
		if request.Topic == "STATUS" {
			scheduler.privateChannel.OutgoingChannel <- broker.ChannelMessage{Topic: "STATUS", Sender: scheduler.name, Content: "OK"}
			log.Debugf("privateChannel: Healthcheck fine ")
		}
		if request.Topic == "SHUTDOWN" {
			return
		}
	}
}

func (scheduler *Scheduler) Stop() {
	scheduler.publicChannel.IncomingChannel <- broker.ChannelMessage{Topic: "SHUTDOWN"}
	if scheduler.wg == nil {
		log.Errorf("Waitgroup cannot be done, if not started")
	} else {
		scheduler.wg.Done()
	}
	log.Infof("%s bot stopped", scheduler.name)
}

func (scheduler *Scheduler) GetServiceName() string {
	return scheduler.name
}

func (scheduler *Scheduler) SetChannels(publicChannel broker.MultiPlexChannel, privatChannel broker.MultiPlexChannel) {
	scheduler.privateChannel = &privatChannel
	scheduler.publicChannel = &publicChannel
}

func InitModule() *Scheduler {
	return &Scheduler{name: "SCHEDULER"}
}
