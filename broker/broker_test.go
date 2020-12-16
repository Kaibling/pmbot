package broker_test

import (
	"pmbot/broker"
	"sync"
	"testing"
)

type testModule struct {
	name           string
	publicChannel  broker.MultiPlexChannel
	privateChannel broker.MultiPlexChannel
}

func (testModule *testModule) Start(*sync.WaitGroup) {
	for {
		message := <-testModule.publicChannel.IncomingChannel
		if message.Topic == "STATUS" {
			testModule.privateChannel.OutgoingChannel <- broker.ChannelMessage{Topic: "STATUS", Sender: testModule.name, Content: "OK"}
		}
	}

}
func (testModule *testModule) Stop() {

}
func (testModule *testModule) GetServiceName() string {
	return testModule.name

}
func (testModule *testModule) SetChannels(publicChannel broker.MultiPlexChannel, privatChannel broker.MultiPlexChannel) {
	testModule.publicChannel = publicChannel
	testModule.privateChannel = privatChannel

}

func TestBrokerAddServiceRight(t *testing.T) {

	brokerInstance := broker.InitBroker()

	testModule := &testModule{name: "testModule"}
	brokerInstance.AddService(testModule)
	brokerInstance.SubscribeTopic("testModule", "TEST")
	go brokerInstance.Start()
	testContent := "TESTCONTENT"
	testModule.publicChannel.OutgoingChannel <- broker.ChannelMessage{Topic: "TEST", Content: testContent}

	response := <-testModule.publicChannel.IncomingChannel
	if response.Content != testContent {
		t.Errorf("Content %s != %s ", response.Content, testContent)
		t.FailNow()
	}
	brokerInstance.Stop()

}

func TestBrokerSubscribeToNonexisitingService(t *testing.T) {

	brokerInstance := broker.InitBroker()

	testModule := &testModule{name: "testModule"}
	brokerInstance.AddService(testModule)
	err := brokerInstance.SubscribeTopic("TEST", "testModule")
	if err == nil {
		t.Errorf("Should be error")
	}
}

func TestBrokerTestStatus(t *testing.T) {

	brokerInstance := broker.InitBroker()

	testModule := &testModule{name: "testModule"}
	brokerInstance.AddService(testModule)
	var wg sync.WaitGroup
	go brokerInstance.Start()
	go testModule.Start(&wg)
	testModule.privateChannel.OutgoingChannel <- broker.NewChannelMessage(testModule.name, "STATUS")

	response := <-testModule.privateChannel.IncomingChannel
	if response.Content != "OK" {
		t.Errorf("Content %s != OK ", response.Content)
	}
	brokerInstance.Stop()

}
func TestBrokerTestStatusWrong(t *testing.T) {

	brokerInstance := broker.InitBroker()

	testModule := &testModule{name: "testModule"}
	brokerInstance.AddService(testModule)
	var wg sync.WaitGroup
	go brokerInstance.Start()
	go testModule.Start(&wg)
	testModule.privateChannel.OutgoingChannel <- broker.ChannelMessage{Topic: "STATUS", Sender: testModule.name}

	response := <-testModule.privateChannel.IncomingChannel
	if response.Content != "O" {
		brokerInstance.Stop()
	} else {
		t.Errorf("Content %s ", response.Content)
		brokerInstance.Stop()
	}

}
