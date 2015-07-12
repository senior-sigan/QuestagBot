package telegram

import "net/http"

// Response is telegram base response envelop
type Response struct {
	Response *http.Response
	Ok       bool        `json:"ok"`
	Result   interface{} `json:"result"`
}

// User struct represents a Telegram user or bot.
type User struct {
	ID        int  `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name,omitempty"`
	Username  string `json:"username,omitempty"`
}

// GroupChat struct represents a group chat.
type GroupChat struct {
	ID    int  `json:"id"`
	Title string `json:"title"`
}

// CommonChat struct is base struct for Chat or User object
type CommonChat struct {
	ID int `json:"id"`
}

// Message struct represents a message.
type Message struct {
	MessageID           int         `json:"message_id"`
	From                User        `json:"from"`
	Date                int         `json:"date"`
	Chat                CommonChat  `json:"chat"`
	ForwardFrom         User        `json:"forward_from,omitempty"`
	ForwardDate         int         `json:"forward_date,omitempty"`
	ReplyToMessage      *Message    `json:"reply_to_message,omitempty"`
	Text                string      `json:"text,omitempty"`
	Audio               Audio       `json:"audio,omitempty"`
	Document            Document    `json:"document,omitempty"`
	Photo               []PhotoSize `json:"photo,omitempty"`
	Sticker             Sticker     `json:"sticker,omitempty"`
	Video               Video       `json:"video,omitempty"`
	Contact             Contact     `json:"contact,omitempty"`
	Location            Location    `json:"location,omitempty"`
	NewChatParticipant  User        `json:"new_chat_participant,omitempty"`
	LeftChatParticipant User        `json:"left_chat_participant,omitempty"`
	NewChatTitle        string      `json:"new_chat_title,omitempty"`
	NewChatPhoto        []PhotoSize `json:"new_chat_photo,omitempty"`
	DeleteChatPhoto     bool        `json:"delete_chat_photo,omitempty"`
	GroupChatCreated    bool        `json:"group_chat_created,omitempty"`
}

// PhotoSize struct represents one size of a photo or a file / sticker thumbnail.
type PhotoSize struct {
	FileID   string `json:"file_id"`
	Width    int    `json:"width"`
	Height   int    `json:"height"`
	FileSize int  `json:"file_size,omitempty"`
}

// Audio struct represents an audio file (voice note).
type Audio struct {
	FileID   string `json:"file_id"`
	Duration int    `json:"duration"`
	MimeType string `json:"mime_type,omitempty"`
	FileSize int  `json:"file_size,omitempty"`
}

// Document struct represents a general file (as opposed to photos and audio files).
type Document struct {
	FileID   string    `json:"file_id"`
	Thumb    PhotoSize `json:"thumb"`
	FileName string    `json:"file_name,omitempty"`
	MimeType string    `json:"mime_type,omitempty"`
	FileSize int     `json:"file_size,omitempty"`
}

// Sticker struct
type Sticker struct {
	FileID   string    `json:"file_id"`
	Width    int       `json:"width"`
	Height   int       `json:"heiht"`
	Thumb    PhotoSize `json:"thumb"`
	FileSize int     `json:"file_size,omitempty"`
}

// Video struct represents a video file.
type Video struct {
	FileID   string    `json:"file_id"`
	Width    int       `json:"width"`
	Height   int       `json:"height"`
	Duration int       `json:"duration"`
	Thumb    PhotoSize `json:"thumb"`
	MimeType string    `json:"mime_type,omitempty"`
	FileSize int     `json:"fil_size,omitempty"`
	Caption  string    `json:"caption,omitempty"`
}

// Contact struct represents a phone contact.
type Contact struct {
	PhoneNumber string `json:"phone_number"`
	FirstName   string `json:"first_name"`
	LastName    string `json:"last_name,omitempty"`
	UserID      string `json:"user_id,omitempty"`
}

// UserProfilePhotos struct represents a user's profile pictures.
type UserProfilePhotos struct {
	TotalCount int         `json:"total_count"`
	Photos     []PhotoSize `json:"photos"`
}

// Location struct represents a point on the map.
type Location struct {
	Longitude float32 `json:"longitude"`
	Latitude  float32 `json:"latitude"`
}

// ReplyKeyboardMarkup struct represents a custom keyboard with reply options.
type ReplyKeyboardMarkup struct {
	Keyboard        [][]string `json:"keyboard"`
	ResizeKeyboard  bool       `json:"resize_keyboard,omitempty"`
	OneTimeKeyboard bool       `json:"one_time_keyboard,omitempty"`
	Selective       bool       `json:"selective,omitempty"`
}

// Update struct represents an incoming update.
type Update struct {
	ID      int   `json:"update_id"`
	Message Message `json:"message,omitempty"`
}
