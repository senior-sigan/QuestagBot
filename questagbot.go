package questagbot

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	"image/jpeg"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"strconv"

	"golang.org/x/net/context"

	hexapic "github.com/blan4/hexapic/core"
	martini "github.com/codegangsta/martini"
	binding "github.com/codegangsta/martini-contrib/binding"
	godotenv "github.com/joho/godotenv"
	appengine "google.golang.org/appengine"
	log "google.golang.org/appengine/log"
	urlfetch "google.golang.org/appengine/urlfetch"
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

// GetID return id as string
func (user User) GetID() string {
	return strconv.Itoa(user.ID)
}

// ReplyKeyboardMarkup is a object represents a custom keyboard
type ReplyKeyboardMarkup struct {
	Keyboard        [][]string `json:"keyboard"`
	ResizeKeyboard  bool       `json:"resize_keyboard"`
	OneTimeKeyboard bool       `json:"one_time_keyboard"`
	Selective       bool       `json:"selective"`
}

// TextMessage is a object for sendMessage API method
type TextMessage struct {
	ChatID                int         `json:"chat_id"`
	Text                  string      `json:"text"`
	DisableWebPagePreview bool        `json:"disable_web_page_preview"`
	ReplyToMessageID      int         `json:"reply_to_message_id"`
	ReplyMarkup           interface{} `json:"reply_markup"`
}

// PhotoMessage is a object for sendPhoto API method
type PhotoMessage struct {
	ChatID           int         `json:"chat_id"`
	Photo            string      `json:"photo"`
	Caption          string      `json:"caption"`
	ReplyToMessageID int         `json:"reply_to_message_id"`
	ReplyMarkup      interface{} `json:"reply_markup"`
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
		//sendMessage(c, apiURL, update, "Hello")
		if err := sendChatAction(c, apiURL, update, "upload_photo"); err != nil {
			log.Criticalf(c, "Can't sendChatAction %v", err)
		}
		if err := sendPhoto(c, os.Getenv("INSTAGRAM_CLIENT_ID"), apiURL, update, ""); err != nil {
			log.Criticalf(c, "Can't sendPhoto %v", err)
		}
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

func sendChatAction(c context.Context, apiURL string, data Update, action string) (err error) {
	httpClient := urlfetch.Client(c)
	query := url.Values{}
	query.Set("action", action)
	query.Add("chat_id", strconv.Itoa(data.Message.Chat.ID))
	url := apiURL + "sendChatAction"
	r, err := http.NewRequest("POST", url, bytes.NewBufferString(query.Encode()))
	if err != nil {
		return
	}
	r.Header.Add("Content-Length", strconv.Itoa(len(query.Encode())))
	resp, err := httpClient.Do(r)
	defer resp.Body.Close()
	log.Infof(c, "sendChatAction: %v %v", url, query.Encode())
	return
}

func sendPhoto(c context.Context, clientdID string, apiURL string, data Update, text string) (err error) {
	httpClient := urlfetch.Client(c)
	hexapicAPI := hexapic.NewSearchApi(clientdID, httpClient)
	hexapicAPI.Count = 4
	var (
		imageQuality = jpeg.Options{Quality: jpeg.DefaultQuality}
		b            bytes.Buffer
		fw           io.Writer
		imgs         []image.Image
	)
	w := multipart.NewWriter(&b)
	if fw, err = w.CreateFormField("chat_id"); err != nil {
		return
	}
	if _, err = fw.Write([]byte(data.Message.Chat.GetID())); err != nil {
		return
	}
	if text != "" {
		if fw, err = w.CreateFormField("caption"); err != nil {
			return
		}
		if _, err = fw.Write([]byte(text)); err != nil {
			return
		}
	}
	if fw, err = w.CreateFormField("reply_markup"); err != nil {
		return
	}
	json, err := keyboardJSON()
	if err != nil {
		return
	}
	if _, err = fw.Write(json); err != nil {
		return
	}
	if fw, err = w.CreateFormFile("photo", "image.jpg"); err != nil {
		return
	}
	imgs = hexapicAPI.SearchByTag("cat")
	img := hexapic.GenerateCollage(imgs, 2, 2)
	if err = jpeg.Encode(fw, img, &imageQuality); err != nil {
		return
	}
	w.Close()

	req, err := http.NewRequest("POST", apiURL+"sendPhoto", &b)
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", w.FormDataContentType())
	req.Header.Add("Content-Length", strconv.Itoa(b.Len()))
	res, err := httpClient.Do(req)
	defer res.Body.Close()
	if err != nil {
		return
	}
	if res.StatusCode != http.StatusOK {
		err = fmt.Errorf("bad status: %s", res.Status)
	}
	return
}

// keyboardJSON create json object for ReplyKeyboardMarkup
func keyboardJSON() (b []byte, err error) {
	km := &ReplyKeyboardMarkup{
		Keyboard:        [][]string{{"cat", "dog"}, {"nya", "chick"}},
		ResizeKeyboard:  true,
		OneTimeKeyboard: true,
		Selective:       false,
	}
	b, err = json.Marshal(km)
	return
}
