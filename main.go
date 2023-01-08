package main

import (
	"github.com/Kaibling/pmbot/lib/config"
	"github.com/Kaibling/pmbot/lib/modules/modulegroup"

	log "github.com/sirupsen/logrus"
)

var version string
var appName = "pmBot"
var buildTime string

func init() {
	log.SetReportCaller(true)
	//log.SetFormatter(&log.JSONFormatter{})
	log.SetLevel(log.DebugLevel)
	//log.SetLevel(log.InfoLevel)

	config.Apply(version)
	log.Infof("Version: %s", version)
	log.Infof("Buildtime: %s", buildTime)
}

func main() {
	mg := modulegroup.NewModuleGroup()
	mg.Start()
}
