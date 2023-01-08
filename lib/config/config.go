package config

import (
	"os"
	"strings"
)

type Config struct {
	Version   string
	Reddit    Reddit
	Discord   Discord
	Variables map[string]string
}

type Reddit struct {
	ClientID   string
	Secret     string
	Username   string
	Password   string
	SubReddits []string
}

func (r Reddit) Set() bool {
	return r.ClientID != "" && r.Secret != "" && r.Username != "" && r.Password != ""
}

type Discord struct {
	Token string
}

func (r Discord) Set() bool {
	return r.Token != ""
}

var Configuration Config

// Apply -
func Apply(version string) {
	c := Config{
		Variables: make(map[string]string),
	}
	r := Reddit{
		ClientID:   os.Getenv("REDDIT_CLIENT_ID"),
		Secret:     os.Getenv("REDDIT_SECRET"),
		Username:   os.Getenv("REDDIT_USERNAME"),
		Password:   os.Getenv("REDDIT_PASSWORD"),
		SubReddits: strings.Split(os.Getenv("REDDIT_SUBREDDITS"), ","),
	}
	if r.Set() {
		c.Reddit = r
	}
	d := Discord{
		Token: os.Getenv("DISCORD_TOKEN"),
	}
	if d.Set() {
		c.Discord = d
	}
	c.Version = version
	Configuration = c
}
