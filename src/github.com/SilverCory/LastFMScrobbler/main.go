package main

import (
	"fmt"
	"github.com/SilverCory/LastFMScrobbler/bot"
	"github.com/SilverCory/LastFMScrobbler/config"
	"github.com/SilverCory/LastFMScrobbler/scrobbler"
	"os"
	"runtime"
)

func main() {

	runtime.GOMAXPROCS(1)

	conf := &config.ScrobblerConfig{}
	conf.Load()

	instance := &bot.Instance{}

	err := instance.Init(conf)
	if err != nil {
		return
	}

	// Start the scrobbler.
	scrobbler.New(conf)

	err = instance.ConnectAndTakeover()
	if err != nil {
		fmt.Println("There was an error connecting to discord.", err)
	}

}
