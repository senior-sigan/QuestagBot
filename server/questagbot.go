package questagbot

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"

	"golang.org/x/net/context"

	//hexapic "github.com/blan4/hexapic/core"
	"github.com/blan4/QuestagBot/telegram"
	"github.com/codegangsta/martini"
	"github.com/codegangsta/martini-contrib/binding"
	"github.com/joho/godotenv"

	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
)

var random = rand.New(rand.NewSource(42))

// Global is struct for saving state
type Global struct {
	InstagramClientID string
	APIURL            string
	Tags              []string
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
	global.APIURL = fmt.Sprintf("https://api.telegram.org/bot%v/", os.Getenv("TELEGRAM_KEY"))

	m := martini.Classic()
	m.Use(appEngine)
	m.Use(martini.Logger())
	m.Get("/", func() string {
		return "Hello world"
	})
	m.Post("/bothook", binding.Bind(telegram.Update{}), func(c context.Context, update telegram.Update, w http.ResponseWriter) string {
		log.Infof(c, "%v", update)
		gamer, err := findOrCreateGamer(update, c)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			log.Errorf(c, "Can't find or create gamer: %v", err)
		}
		log.Infof(c, "Gamer : %v", gamer.ChatID)
		//handleComand(update)
		//sendMessage(c, apiURL, update, "Hello")
		// if err := sendChatAction(c, update, "upload_photo"); err != nil {
		// 	log.Criticalf(c, "Can't sendChatAction %v", err)
		// }
		// if err := sendPhoto(c, update, ""); err != nil {
		// 	log.Criticalf(c, "Can't sendPhoto %v", err)
		// }
		return strconv.Itoa(update.ID)
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

// func generateImage() {
// 	hexapicAPI := hexapic.NewSearchApi(global.InstagramClientID, httpClient)
// 	hexapicAPI.Count = 4
// 	imgs := hexapicAPI.SearchByTag(question.Answer)
// 	img := hexapic.GenerateCollage(imgs, 2, 2)
// 	question := state.NextQuestion()
// }

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

// NextQuestion return next question
func (gamer *Gamer) NextQuestion() (question Question) {
	question = gamer.Questions[gamer.CurrentQuestion]
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
