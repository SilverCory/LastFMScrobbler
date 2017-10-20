package scrobbler

import (
	"time"

	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io/ioutil"
	"net/http"
	"strconv"
	"sync"

	"mime"

	"github.com/SilverCory/LastFMScrobbler/bot"
	"github.com/SilverCory/LastFMScrobbler/config"
	"github.com/bwmarrin/discordgo"
)

func New(scorbblerConfig *config.ScrobblerConfig) {

	LASTFMKEY = scorbblerConfig.LastFMAPIKey
	LASTFMUSER = scorbblerConfig.LastFMUser

	scrobbleModule := &ScrobbleModule{
		uploadMux: &sync.Mutex{},
	}

	manager, err := NewManager(scorbblerConfig.DiscordWebToken, scorbblerConfig.DiscordAppID)
	if err != nil {
		fmt.Println("Unable to start the asset manager? Token wrong?", err)
		return
	}

	scrobbleModule.Assets = manager
	scrobbleModule.PruneAlbumCovers()

	go func() {
		for {
			scrobbleModule.Scrobble()
			time.Sleep(10 * time.Second)
		}
	}()

}

type ScrobbleModule struct {
	Assets       *AssetManager
	lastId       string
	currentId    string
	uploadMux    *sync.Mutex
	lastString   string
	lastDuration time.Duration
	TimeThen     time.Time
}

func (s *ScrobbleModule) UploadCover(url string) {
	s.uploadMux.Lock()

	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("Error downloading image!", err)
		s.uploadMux.Unlock()
		return
	}

	if resp.StatusCode < 200 || resp.StatusCode > 209 {
		fmt.Println("Error in Downloading!", "Status code not 200, instead : "+strconv.Itoa(resp.StatusCode)+", status: "+resp.Status)
		s.uploadMux.Unlock()
		return
	}

	defer resp.Body.Close()
	imageBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading image body.", err)
		s.uploadMux.Unlock()
		return
	}

	_, format, err := image.Decode(bytes.NewBuffer(imageBody))
	if err != nil {
		fmt.Println("Error decoding image.", err)
		s.uploadMux.Unlock()
		return
	}

	data := mime.TypeByExtension("." + format)
	img64 := base64.StdEncoding.EncodeToString(imageBody)
	asset, err := s.Assets.AddAsset("album_cover", "data:"+data+";base64,"+img64, 2)
	if err != nil {
		fmt.Println("Error uploading asset!", err)
		s.uploadMux.Unlock()
		return
	}

	s.currentId = asset.ID
	s.uploadMux.Unlock()
}

func (s *ScrobbleModule) Scrobble() {

	if bot.Bot == nil || bot.Bot.DiscordClient == nil {
		return
	}

	resp, err := http.Get("https://ws.audioscrobbler.com/2.0/?method=user.getrecenttracks&user=" + LASTFMUSER + "&api_key=" + LASTFMKEY + "&format=json&limit=1")
	if err != nil {
		fmt.Println("Error in Scrobble!", err)
		return
	}

	if resp.StatusCode < 200 || resp.StatusCode > 209 {
		fmt.Println("Error in Scrobble!", "Status code not 200, instead : "+strconv.Itoa(resp.StatusCode)+", status: "+resp.Status)
		return
	}

	defer resp.Body.Close()
	jsonArr, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error in Scrobble!", err)
		return
	}

	scrobbleResponse := &ScrobbleResponse{}
	if err := json.Unmarshal(jsonArr, scrobbleResponse); err != nil {
		fmt.Println("Error in Scrobble!", err)
		return
	}

	recentTracks := scrobbleResponse.Tracks
	track := recentTracks.FindNowPlaying()
	if track == nil && time.Now().After(s.TimeThen.Add(s.lastDuration+(5*time.Second))) {
		bot.Bot.DiscordClient.UpdateStatus(0, "")
		return
	}

	track.LoadDuration()
	compareString := track.Name + "^^^" + track.Artist.Text
	if s.lastString != compareString {
		if s.currentId != bot.Bot.Conf.DiscordDefaultImageID {
			s.lastId = s.currentId
		}
		s.currentId = bot.Bot.Conf.DiscordDefaultImageID
		if url := track.FindImageURL(); url != "" {
			s.UploadCover(url)
		} else {
			fmt.Println("URL was empty!")
		}
		s.Assets.RemoveAsset(s.lastId)
		s.TimeThen = time.Now()
	}

	s.lastString = compareString

	bot.Bot.DiscordClient.UpdateStatusRaw(0, &discordgo.Game{
		State:         "M O O S I C",
		Details:       track.Artist.Text,
		Name:          track.Name,
		Type:          0,
		URL:           "",
		ApplicationID: bot.Bot.Conf.DiscordAppID,
		Assets: discordgo.Assets{
			LargeImageID: s.currentId,
			SmallImageID: bot.Bot.Conf.DiscordSmallImageID,
			LargeText:    track.Album.Text,
			SmallText:    "cory.red",
		},
		TimeStamps: discordgo.TimeStamps{
			StartTimestamp: uint(s.TimeThen.Unix()) - 5,
			EndTimestamp:   uint(s.TimeThen.Add(track.Duration).Unix()),
		},
	})

}

func (s *ScrobbleModule) PruneAlbumCovers() {

	assets, err := s.Assets.GetAssetsWithName("album_cover")
	if err != nil {
		fmt.Println("Error pruning :(", err)
		return
	}

	for _, v := range *assets {
		s.Assets.RemoveAsset(v.ID)
	}

}
