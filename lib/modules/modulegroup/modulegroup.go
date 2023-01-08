package modulegroup

import (
	"os"
	"os/signal"
	"sync"
	"syscall"

	log "github.com/sirupsen/logrus"

	"github.com/Kaibling/pmbot/lib/broker"
	"github.com/Kaibling/pmbot/lib/config"
	"github.com/Kaibling/pmbot/lib/modules/discord"
	"github.com/Kaibling/pmbot/lib/modules/reddit"
	//"github.com/Kaibling/pmbot/lib/modules/reddit"
)

type ModuleGroup struct {
	modules []broker.Module
}

func NewModuleGroup() (mg *ModuleGroup) {
	mg = &ModuleGroup{}
	//reddit
	if config.Configuration.Reddit.Set() {
		mg.modules = append(mg.modules, reddit.InitModule(
			config.Configuration.Reddit.Username,
			config.Configuration.Reddit.ClientID,
			config.Configuration.Reddit.Secret,
			config.Configuration.Reddit.Password,
			config.Configuration.Reddit.SubReddits))
	}
	//discord
	if config.Configuration.Discord.Set() {
		mg.modules = append(mg.modules, discord.InitModule(config.Configuration.Discord.Token))
	}
	//Scheduler
	//modules = append(modules, scheduler.InitModule())
	return
}

func (mg *ModuleGroup) Start() {
	if len(mg.modules) == 0 {
		log.Info("no modules configured")
		return
	}
	sc := make(chan os.Signal, 1)

	broker := broker.NewBroker()
	//brokerInstance.SubscribeTopic("DISCORD", "REDDIT")

	//START MODULES
	var wg sync.WaitGroup
	wg.Add(len(mg.modules))
	for _, module := range mg.modules {
		broker.AddModule(module)
		go module.Start(&wg)
	}
	go broker.Start()
	log.Infoln("All modules loaded. Press CTRL-C to exit.")
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	for {
		msg1 := <-sc
		broker.Stop()
		for _, module := range mg.modules {
			go module.Stop()
		}
		wg.Wait()
		log.Debugln("Stopp die scheiße", msg1)
		break
	}
}
