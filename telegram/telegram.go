package telegram

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	"image/jpeg"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"net/url"
	"strconv"
)

// BotURL is url to telegram bot api
const BotURL string = "https://api.telegram.org/bot"

// Telegram is struct to store client data
type Telegram struct {
	Key    string
	client *http.Client
	botURL string
}

// NewTelegram is factory method for Telegram struct
func NewTelegram(httpClient *http.Client, key string) *Telegram {
	return &Telegram{
		Key:    key,
		client: httpClient,
		botURL: fmt.Sprintf("https://api.telegram.org/bot%v/", key),
	}
}

// do is helper function to send requests
func (telegram Telegram) do(req *http.Request, v interface{}) (*Response, error) {
	log.Printf("Sending request: %v", req)
	resp, err := telegram.client.Do(req)
	if err != nil {
		return &Response{}, err
	}

	defer resp.Body.Close()

	r := &Response{Response: resp}
	if v != nil {
		r.Result = v
		err = json.NewDecoder(resp.Body).Decode(r)
	}

	return r, err
}

// GetMe A simple method for testing your bot's auth token.
// Requires no parameters.
// Returns basic information about the bot in form of a User object.
func (telegram Telegram) GetMe() (*User, error) {
	uri := telegram.botURL + "getMe"
	req, err := http.NewRequest("GET", uri, bytes.NewBufferString(""))
	if err != nil {
		return &User{}, err
	}

	user := new(User)
	_, err = telegram.do(req, user)

	return user, err
}

// SendMessage - Use this method to send text messages
func (telegram Telegram) SendMessage(chatID int, text string, disableWebPagePreview bool, replyToMessageID int, replyMarkup *ReplyKeyboardMarkup) (*Message, error) {
	msg := new(Message)
	uri := telegram.botURL + "sendMessage"
	query := url.Values{}
	query.Set("text", text)
	query.Add("chat_id", strconv.Itoa(chatID))
	//query.Add("disable_web_page_preview", strconv.FormatBool(disableWebPagePreview))
	if replyToMessageID != 0 {
		query.Add("reply_to_message_id", strconv.Itoa(replyToMessageID))
	}
	if replyMarkup != nil {
		bReplyMarkup, err := json.Marshal(replyMarkup)
		if err != nil {
			return msg, err
		}
		query.Add("reply_markup", string(bReplyMarkup))
	}
	uri = uri + "?" + query.Encode()
	req, err := http.NewRequest("POST", uri, bytes.NewBufferString(""))
	req.Header.Add("Content-Length", strconv.Itoa(len(query.Encode())))
	_, err = telegram.do(req, msg)
	return msg, err
}

// SendPhoto - Use this method to send photos.
func (telegram Telegram) SendPhoto(chatID int, photo image.Image, caption string, replyToMessageID int, replyMarkup *ReplyKeyboardMarkup) (*Message, error) {
	if err := telegram.sendChatAction(chatID, "upload_photo"); err != nil {
		return nil, err
	}
	var (
		uri          = telegram.botURL + "sendPhoto"
		imageQuality = jpeg.Options{Quality: jpeg.DefaultQuality}
		b            bytes.Buffer
		fw           io.Writer
		msg          = new(Message)
		err          error
	)
	w := multipart.NewWriter(&b)
	if fw, err = w.CreateFormField("chat_id"); err != nil {
		return msg, err
	}
	if _, err = fw.Write([]byte(strconv.Itoa(chatID))); err != nil {
		return msg, err
	}
	if caption != "" {
		if fw, err = w.CreateFormField("caption"); err != nil {
			return msg, err
		}
		if _, err = fw.Write([]byte(caption)); err != nil {
			return msg, err
		}
	}
	if replyToMessageID != 0 {
		if fw, err = w.CreateFormField("reply_to_message_id"); err != nil {
			return msg, err
		}
		if _, err = fw.Write([]byte(strconv.Itoa(replyToMessageID))); err != nil {
			return msg, err
		}
	}
	if replyMarkup != nil {
		keyboard, err := json.Marshal(replyMarkup)
		if err != nil {
			return msg, err
		}
		if fw, err = w.CreateFormField("reply_markup"); err != nil {
			return msg, err
		}
		if _, err = fw.Write([]byte(keyboard)); err != nil {
			return msg, err
		}
	}
	if fw, err = w.CreateFormFile("photo", "image.jpg"); err != nil {
		return msg, err
	}
	if err = jpeg.Encode(fw, photo, &imageQuality); err != nil {
		return msg, err
	}
	w.Close()

	req, err := http.NewRequest("POST", uri, &b)
	if err != nil {
		return msg, err
	}
	req.Header.Set("Content-Type", w.FormDataContentType())
	req.Header.Add("Content-Length", strconv.Itoa(b.Len()))

	_, err = telegram.do(req, msg)
	return msg, err
}

func (telegram Telegram) sendChatAction(chatID int, action string) (err error) {
	query := url.Values{}
	query.Set("action", action)
	query.Add("chat_id", strconv.Itoa(chatID))
	uri := telegram.botURL + "sendChatAction" + "?" + query.Encode()
	req, err := http.NewRequest("POST", uri, bytes.NewBufferString(""))
	if err != nil {
		return
	}
	req.Header.Add("Content-Length", strconv.Itoa(len(query.Encode())))
	_, err = telegram.do(req, nil)
	return
}
