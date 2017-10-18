package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

type ScrobblerConfig struct {
	LastFMAPIKey          string `json:"last_fmapi_key"`
	LastFMUser            string `json:"last_fm_user"`
	DiscordAppID          string `json:"discord_app_id"`
	DiscordWebToken       string `json:"discord_web_token"`
	DiscordBotToken       string `json:"discord_bot_token"`
	DiscordSmallImageID   string `json:"discord_small_image_id"`
	DiscordDefaultImageID string `json:"discord_default_image_id"`
}

func (co *ScrobblerConfig) Load() {
	if _, err := os.Stat("./config.json"); os.IsNotExist(err) {
		co.Save()
		fmt.Println("The default configuration has been saved. Please edit this and restart!")
		os.Exit(0)
		return
	} else {
		data, err := ioutil.ReadFile("./config.json")
		if err != nil {
			fmt.Println("There was an error loading the config!", err)
			return
		}

		err = json.Unmarshal(data, co)
		if err != nil {
			co.LastFMAPIKey = "lastfmapikey"
			co.LastFMUser = "coryory"
			co.DiscordAppID = "368924690946850817"
			co.DiscordWebToken = "web token"
			co.DiscordDefaultImageID = "368982525487480832"
			co.DiscordSmallImageID = "368981253099225098"
			fmt.Println("There was an error loading the config!", err)
			return
		}
	}
}

func (co *ScrobblerConfig) Save() error {
	data, err := json.MarshalIndent(co, "", "\t")

	if err != nil {
		fmt.Println("There was an error saving the config!", err)
		return err
	}

	err = ioutil.WriteFile("./config.json", data, 0644)
	if err != nil {
		fmt.Println("There was an error saving the config!", err)
		return err
	}

	return nil

}
