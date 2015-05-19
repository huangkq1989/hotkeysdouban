package douban

import "fmt"
import "net/http"
import "net/url"
import "io/ioutil"
import "encoding/json"
import "net/http/cookiejar"
import "strconv"

const (
	DOUBAN_APP_NAME     = "radio_desktop_win"
	DOUBAN_APP_VERSION  = "100"
	DOUBAN_BASE_URI     = "http://www.douban.com"
	DOUBAN_SIGNIN       = "/j/app/login"
	DOUBAN_GET_CHANNELS = "/j/app/radio/channels"
	DOUBAN_OPERATE_SONG = "/j/app/radio/people"
)

type Douban struct {
	cookie      []*http.Cookie
	version     string
	appName     string
	baseUrl     string
	loginResult LoginResult
}

type LoginResult struct {
	Userid   string `json:"user_id"`
	Err      string `json:"err"`
	Token    string `json:"token"`
	Expire   string `json:"expire"`
	Username string `json:"user_name"`
	Email    string `json:"email"`
	R        int    `json:"r"`
}

func (douban *Douban) Signin(user, passwd string) bool {
	signinUrl := douban.baseUrl + DOUBAN_SIGNIN
	resp, err := http.PostForm(signinUrl, url.Values{"email": {user},
		"password": {passwd},
		"app_name": {douban.appName},
		"version":  {douban.version}})
	if err != nil {
		fmt.Println("Fail to Login:", err)
		return false
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Fail to read response", err)
		return false
	} else {
		result := &LoginResult{}
		err := json.Unmarshal(body, &result)
		if err != nil {
			fmt.Println("Login Failed:", err)
		}
		douban.cookie = resp.Cookies()
		douban.loginResult = *result
	}
	return true
}

func (douban *Douban) GetChannels() map[string]string {
	channelsMap := make(map[string]string)
	channels_url := douban.baseUrl + DOUBAN_GET_CHANNELS
	resp, err := http.PostForm(channels_url, url.Values{})
	if err != nil {
		fmt.Println("Fail to get channels", err)
		return channelsMap
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Fail to read response", err)
		return channelsMap
	} else {
		var result map[string]interface{}
		json.Unmarshal(body, &result)
		channels := result["channels"].([]interface{})
		for i := 0; i < len(channels); i++ {
			name := channels[i].(map[string]interface{})["name"].(string)
			switch channels[i].(map[string]interface{})["channel_id"].(type) {
			case float64:
				id := channels[i].(map[string]interface{})["channel_id"].(float64)
				channelId := strconv.FormatFloat(id, 'f', 0, 64)
				channelsMap[channelId] = name
				fmt.Println(channelId, "\t=>\t", name)
			case string:
				channelId := channels[i].(map[string]interface{})["channel_id"].(string)
				channelsMap[channelId] = name
				fmt.Println(channelId, "\t=>\t", name)
			}
		}
	}
	return channelsMap
}

type Song struct {
	Title  string
	Artist string
	Url    string
	SongId string
}

func (douban *Douban) GetSongList(channelId string) []Song {
	jar, _ := cookiejar.New(nil)
	client := http.Client{nil, nil, jar, 0}

	songListUrl := douban.baseUrl + DOUBAN_OPERATE_SONG

	payload := url.Values{"app_name": {douban.appName},
		"version": {douban.version},
		"user_id": {douban.loginResult.Userid},
		"expire":  {douban.loginResult.Expire},
		"token":   {douban.loginResult.Token},
		"channel": {channelId},
		"type":    {"n"}}

	resp, err := client.PostForm(songListUrl, payload)
	if err != nil {
		fmt.Println("Fail to get song list:", err)
		return nil
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)

	songList := []Song{}
	if err != nil {
		fmt.Println("Fail to read response of get song list:", err)
		return nil
	} else {
		var result map[string]interface{}
		json.Unmarshal(body, &result)
		song := result["song"].([]interface{})
		for i := 0; i < len(song); i++ {
			title := song[i].(map[string]interface{})["title"].(string)
			artist := song[i].(map[string]interface{})["artist"].(string)
			url := song[i].(map[string]interface{})["url"].(string)
			songId := song[i].(map[string]interface{})["sid"].(string)

			song := Song{title, artist, url, songId}
			songList = append(songList, song)
		}
	}
	return songList
}

func (douban *Douban) songApi(channelId, songId, operateType string) bool {
	payload := url.Values{"app_name": {douban.appName},
		"version": {douban.version},
		"user_id": {douban.loginResult.Userid},
		"expire":  {douban.loginResult.Expire},
		"token":   {douban.loginResult.Token},
		"channel": {channelId},
		"sid":     {songId},
		"h":       {""},
		"type":    {operateType}}
	apiUrl := douban.baseUrl + DOUBAN_OPERATE_SONG
	jar, _ := cookiejar.New(nil)
	client := http.Client{nil, nil, jar, 0}
	_, err := client.PostForm(apiUrl, payload)

	if err != nil {
		fmt.Println("Fail to request", err)
		return false
	}
	return true
}

func (douban *Douban) RateSong(channelId, songId string) bool {
	RATE_IT := "r"
	return douban.songApi(channelId, songId, RATE_IT)
}

func (douban *Douban) UnrateSong(channelId, songId string) bool {
	UNRATE_IT := "u"
	return douban.songApi(channelId, songId, UNRATE_IT)
}

func (douban *Douban) ByeSong(channelId, songId string) bool {
	BYE_IT := "b"
	return douban.songApi(channelId, songId, BYE_IT)
}
