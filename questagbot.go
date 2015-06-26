package questagbot

import (
	"bytes"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strconv"

	"golang.org/x/net/context"

	martini "github.com/codegangsta/martini"
	binding "github.com/codegangsta/martini-contrib/binding"
	godotenv "github.com/joho/godotenv"
	appengine "google.golang.org/appengine"
	log "google.golang.org/appengine/log"
	urlfetch "google.golang.org/appengine/urlfetch"

	// hexapic "github.com/blan4/hexapic/core"
)

// Update struct from telegram Webhook API
type Update struct {
	ID      int         `json:"update_id"`
	Message MessageMeta `json:"message"`
}

// MessageMeta struct from telegram Webhook API
type MessageMeta struct {
	ID   int    `json:"message_id"`
	Date int    `json:"date"`
	Text string `json:"text"`
	From User   `json:"from"`
	Chat User   `json:"chat"`
}

// User struct from telegram Webhook API
type User struct {
	ID        int    `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Username  string `json:"username"`
}

func appEngine(c martini.Context, r *http.Request) {
	c.Map(appengine.NewContext(r))
}

func init() {
	godotenv.Load("secrets.env")
	apiURL := fmt.Sprintf("https://api.telegram.org/bot%v/", os.Getenv("TELEGRAM_KEY"))

	m := martini.Classic()
	m.Use(appEngine)
	m.Use(martini.Logger())
	m.Get("/", func() string {
		return "Hello world"
	})
	m.Post("/bothook", binding.Bind(Update{}), func(c context.Context, update Update) string {
		log.Infof(c, "%v", update)
		sendMessage(c, apiURL, update, "Hello")
		return strconv.Itoa(update.ID)
	})
	http.Handle("/", m)
}

func sendMessage(c context.Context, apiURL string, data Update, text string) {
	httpClient := urlfetch.Client(c)
	query := url.Values{}
	query.Set("text", "Hello")
	query.Add("chat_id", strconv.Itoa(data.Message.Chat.ID))
	url := apiURL + "sendMessage"
	log.Infof(c, "%v", url)
	r, err := http.NewRequest("POST", url, bytes.NewBufferString(query.Encode()))
	if err != nil {
		log.Criticalf(c, "%v", err)
	}
	r.Header.Add("Content-Length", strconv.Itoa(len(query.Encode())))
	resp, err := httpClient.Do(r)
	defer resp.Body.Close()
	if err != nil {
		log.Criticalf(c, "%v", err)
	}
	log.Infof(c, "%v", resp)
}
