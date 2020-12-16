package configuration

type Config struct {
	Reddit       RedditConfig
	DiscordToken string
	Variables    map[string]string
}
type RedditConfig struct {
	ClientID string
	Secret   string
	Username string
	Password string
}

var Configuration Config

//Apply -
func Apply(config Config) {
	Configuration = config

}
