package questagbot

import (
	"net/http"
	"strconv"

	"golang.org/x/net/context"

	martini "github.com/codegangsta/martini"
	binding "github.com/codegangsta/martini-contrib/binding"
	appengine "google.golang.org/appengine"
	log "google.golang.org/appengine/log"
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
	m := martini.Classic()
	m.Use(appEngine)
	m.Use(martini.Logger())
	m.Get("/", func() string {
		return "Hello world"
	})
	m.Post("/bothook", binding.Bind(Update{}), func(c context.Context, update Update) string {
		log.Infof(c, "%v", update)
		return strconv.Itoa(update.ID)
	})
	http.Handle("/", m)
}
