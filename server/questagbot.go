package questagbot

import (
	"encoding/json"
	"image"
	"math/rand"
	"net/http"
	"os"
	"strings"

	"golang.org/x/net/context"

	"github.com/blan4/QuestagBot/telegram"
	hexapic "github.com/blan4/hexapic/core"
	"github.com/codegangsta/martini"
	"github.com/codegangsta/martini-contrib/binding"
	"github.com/joho/godotenv"

	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/urlfetch"
)

var random = rand.New(rand.NewSource(42))

// Global is struct for saving state
type Global struct {
	InstagramClientID string
	Tags              []string
	TelegramKey       string
}

// Question is struct to store question object
type Question struct {
	Answer   string   `json:"answer"`
	Variants []string `json:"variants"`
}

// Gamer is object to store in appengine datastore
type Gamer struct {
	ChatID          int        `json:"chat_id"`
	Questions       []Question `json:"questions"`
	CurrentQuestion int        `json:"current_question"`
	RightAnswers    int        `json:"right_answers"`
	WrongAnswers    int        `json:"wrong_answers"`
}

// GamerData is wrapper for appengine data store
type GamerData struct {
	GamerBlob string
	Gamer     *Gamer `datastore:"-"`
}

// Load is google store Question struct loader
func (data *GamerData) Load(p []datastore.Property) error {
	if err := datastore.LoadStruct(data, p); err != nil {
		return err
	}
	data.Gamer = new(Gamer)
	return json.Unmarshal([]byte(data.GamerBlob), data.Gamer)
}

// Save is google store Question struct saver
func (data *GamerData) Save() ([]datastore.Property, error) {
	blob, err := json.Marshal(data.Gamer)
	if err != nil {
		panic(err)
	}

	return []datastore.Property{
		datastore.Property{
			Name:    "GamerBlob",
			Value:   string(blob),
			NoIndex: true,
		},
	}, nil
}

func findGamer(c context.Context, id int64) (*Gamer, error) {
	data := new(GamerData)
	key := datastore.NewKey(c, "Gamer", "", id, nil)
	if err := datastore.Get(c, key, data); err != nil {
		return new(Gamer), err
	}
	return data.Gamer, nil
}

func saveGamer(c context.Context, gamer *Gamer) (err error) {
	data := new(GamerData)
	data.Gamer = gamer
	key := datastore.NewKey(c, "Gamer", "", int64(gamer.ChatID), nil)
	_, err = datastore.Put(c, key, data)
	return
}

func appEngine(c martini.Context, r *http.Request) {
	c.Map(appengine.NewContext(r))
}

var global Global

func init() {
	godotenv.Load("secrets.env")
	global.Tags = strings.Split(os.Getenv("TAGS"), ",")
	global.InstagramClientID = os.Getenv("INSTAGRAM_CLIENT_ID")
	global.TelegramKey = os.Getenv("TELEGRAM_KEY")

	m := martini.Classic()
	m.Use(appEngine)
	m.Use(martini.Logger())
	m.Get("/", func() string {
		return "Hello world"
	})
	m.Post("/bothook", binding.Bind(telegram.Update{}), func(c context.Context, update telegram.Update, w http.ResponseWriter) {
		httpClient := urlfetch.Client(c)
		tele := telegram.NewTelegram(httpClient, global.TelegramKey)
		log.Infof(c, "%v", update)
		gamer, err := findOrCreateGamer(update, c)
		defer saveGamer(c, gamer)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			log.Errorf(c, "Can't find or create gamer: %v", err)
			return
		}
		log.Infof(c, "Gamer : %v", gamer.ChatID)

		if strings.Index(update.Message.Text, "/start") == 0 {
			log.Infof(c, "Start game with %v, %v", gamer.ChatID, update.Message.From.Username)
			gamer.handleStart()
			tele.SendPhoto(update.Message.Chat.ID, generateImage(gamer.NextQuestion(), httpClient), "", 0, gamer.GetKeyboard())
			return
		}
		if strings.Index(update.Message.Text, "/stop") == 0 {
			log.Infof(c, "Stop game with %v, %v", gamer.ChatID, update.Message.From.Username)
			gamer.handleStop()
			tele.SendMessage(update.Message.Chat.ID, "Game over", true, 0, nil)
			return
		}
		if strings.Index(update.Message.Text, "/top") == 0 {
			log.Infof(c, "Show top for %v, %v", gamer.ChatID, update.Message.From.Username)
			gamer.handleTop()
			tele.SendMessage(update.Message.Chat.ID, "Top 10 gamers", true, 0, nil)
			return
		}
		if strings.Index(update.Message.Text, "/help") == 0 {
			log.Infof(c, "Show help for %v, %v", gamer.ChatID, update.Message.From.Username)
			tele.SendMessage(update.Message.Chat.ID, "Todo: help page", true, 0, nil)
			return
		}
		if gamer.isPlaying() {
			log.Infof(c, "Get answer from %v, %v on question %v", gamer.ChatID, update.Message.From.Username, gamer.GetCurrentQuestion())
			if gamer.handleAnswer(update.Message.Text) {
				log.Infof(c, "Right answer, gamer: %v, %v", gamer.ChatID, update.Message.From.Username)
				tele.SendMessage(update.Message.Chat.ID, "👍 Right!", true, 0, nil)
			} else {
				log.Infof(c, "Wrong answer, gamer: %v, %v", gamer.ChatID, update.Message.From.Username)
				tele.SendMessage(update.Message.Chat.ID, "😕 Wrong, "+gamer.GetCurrentQuestion().Answer, true, 0, nil)
			}
			tele.SendPhoto(update.Message.Chat.ID, generateImage(gamer.NextQuestion(), httpClient), "", 0, gamer.GetKeyboard())
			return
		}
		log.Infof(c, "Show help for %v, %v", gamer.ChatID, update.Message.From.Username)
		tele.SendMessage(update.Message.Chat.ID, "Todo: help page", true, 0, nil)
	})
	http.Handle("/", m)
}

func findOrCreateGamer(update telegram.Update, c context.Context) (gamer *Gamer, err error) {
	chatID := update.Message.Chat.ID
	if gamer, err = findGamer(c, int64(chatID)); err != nil {
		log.Infof(c, "Can't find gamer object for this chat: %v, %v", chatID, err)
		gamer.handleStart()
		gamer.ChatID = chatID
		if err := saveGamer(c, gamer); err != nil {
			log.Errorf(c, "Can't store in DB new gamer %v: %v", gamer, err)
			return nil, err
		}
		log.Infof(c, "Saved: %v", gamer.ChatID)
	} else {
		log.Infof(c, "Find gamer with id %v", chatID)
	}
	return gamer, nil
}

func generateImage(question Question, httpClient *http.Client) (img image.Image) {
	hexapicAPI := hexapic.NewSearchApi(global.InstagramClientID, httpClient)
	hexapicAPI.Count = 4
	imgs := hexapicAPI.SearchByTag(question.Answer)
	img = hexapic.GenerateCollage(imgs, 2, 2)
	return
}

// GetKeyboard helper to generate keyboard markup
func (gamer *Gamer) GetKeyboard() (kb *telegram.ReplyKeyboardMarkup) {
	question := gamer.GetCurrentQuestion()
	kb.OneTimeKeyboard = true
	kb.Keyboard = [][]string{
		[]string{question.Variants[0], question.Variants[1]},
		[]string{question.Variants[2], question.Variants[3]},
	}
	return nil
}

// GetCurrentQuestion is helper method to get current question
func (gamer *Gamer) GetCurrentQuestion() Question {
	return gamer.Questions[gamer.CurrentQuestion]
}
func (gamer *Gamer) handleStart() {
	gamer.Questions = generateQuestionsQueue()
	gamer.CurrentQuestion = 0
}
func (gamer *Gamer) handleStop() {
	gamer.Questions = nil
	gamer.CurrentQuestion = 0
}
func (gamer *Gamer) handleTop()  {}
func (gamer *Gamer) handleHelp() {}
func (gamer *Gamer) handleAnswer(answer string) (isRight bool) {
	currentQuestion := gamer.GetCurrentQuestion()
	if currentQuestion.Answer == answer {
		gamer.RightAnswers++
		isRight = true
	} else {
		gamer.WrongAnswers++
		isRight = false
	}

	return
}

func (gamer *Gamer) isPlaying() bool {
	return gamer.Questions != nil
}

// NextQuestion return next question
func (gamer *Gamer) NextQuestion() (question Question) {
	question = gamer.GetCurrentQuestion()
	gamer.CurrentQuestion++
	if gamer.CurrentQuestion == len(global.Tags) {
		gamer.CurrentQuestion = 0
	}
	return
}

func generateQuestionsQueue() []Question {
	tags := global.Tags
	answers := random.Perm(len(tags))
	questions := make([]Question, 0, len(tags))
	for answer := range answers {
		variants := perm(4, len(tags), answer)

		variantsStr := make([]string, len(variants))
		for i, variant := range variants {
			variantsStr[i] = tags[variant]
		}

		question := Question{
			Answer:   tags[answer],
			Variants: variantsStr,
		}

		questions = append(questions, question)
	}

	return questions
}

func perm(size int, limit int, exclude int) []int {
	array := make([]int, size)
	i := 0
	for i < size-1 {
		r := rand.Intn(limit)
		if r != exclude {
			array[i] = r
			i++
		}
	}
	array[size-1] = exclude
	return array
}
