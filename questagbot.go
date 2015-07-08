package questagbot

import (
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"

	//hexapic "github.com/blan4/hexapic/core"
	"github.com/blan4/QuestagBot/telegram"

	"github.com/codegangsta/martini"
	"github.com/codegangsta/martini-contrib/binding"
	"github.com/joho/godotenv"
	"golang.org/x/net/context"
	"google.golang.org/appengine"
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
	Answer   string
	Variants []string
}

// Gamer is object to store in appengine datastore
type Gamer struct {
	ChatID          int
	Questions       []Question
	CurrentQuestion int
	RightAnswers    int
	WrongAnswers    int
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
	m.Post("/bothook", binding.Bind(telegram.Update{}), func(c context.Context, update telegram.Update) string {
		log.Infof(c, "%v", update)
		findOrCreateGamer(update)
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

func findOrCreateGamer(update telegram.Update) *Gamer {
	return nil
}

// func generateImage() {
// 	hexapicAPI := hexapic.NewSearchApi(global.InstagramClientID, httpClient)
// 	hexapicAPI.Count = 4
// 	imgs := hexapicAPI.SearchByTag(question.Answer)
// 	img := hexapic.GenerateCollage(imgs, 2, 2)
// 	question := state.NextQuestion()
// }

// GetCurrentQuestion is helper method to get current question
func (gamer Gamer) GetCurrentQuestion() Question {
	return gamer.Questions[gamer.CurrentQuestion]
}
func (gamer Gamer) handleStart() {
	gamer.Questions = generateQuestionsQueue()
	gamer.CurrentQuestion = 0
}
func (gamer Gamer) handleStop() {
	gamer.Questions = nil
	gamer.CurrentQuestion = 0
}
func (gamer Gamer) handleTop()  {}
func (gamer Gamer) handleHelp() {}
func (gamer Gamer) handleAnswer(answer string) (isRight bool) {
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
func (gamer Gamer) NextQuestion() (question Question) {
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
