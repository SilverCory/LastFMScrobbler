package bot

import (
	"fmt"
	"github.com/SilverCory/LastFMScrobbler/config"
	"github.com/bwmarrin/discordgo"
	"os"
	"os/signal"
	"syscall"
)

var Bot *Instance

type Instance struct {
	DiscordClient *discordgo.Session
	Conf          *config.ScrobblerConfig
}

func (i *Instance) Init(conf *config.ScrobblerConfig) error {

	var err error

	i.Conf = conf
	i.DiscordClient, err = discordgo.New(conf.DiscordBotToken)
	if err != nil {
		fmt.Println("Unable to log into discord.", err)
		return err
	}

	Bot = i
	return nil

}

func (i *Instance) ConnectAndTakeover() error {

	err := i.DiscordClient.Open()
	if err != nil {
		return err
	}

	// Wait here until CTRL-C or other term signal is received.
	fmt.Println("Bot is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	i.DiscordClient.Close()

	return nil

}
