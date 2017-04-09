package checkup

import (
	"log"
	"math/rand"
	"strings"
	"time"

	"github.com/nlopes/slack"
)

type EndpointState struct {
	LastChecked int64
	LastStatus  string
}

// SlackNotifier is the main struct consist of all the sub component including slack api, real-time messaing api and face detector
type SlackNotifier struct {
	ID             string
	RTM            *slack.RTM
	SlackApi       *slack.Client
	EndpointStates []EndpointState
	ChannelID      string
}

const (
	helpText = "How can I help you? mention me and upload your photos, i will do some magic"
)

var (
	greetingPattern  = "hi bot|hello bot"
	greetingPrefixes = []string{"Hi", "Hello", "Howdy", "Wazzzup", "Hey"}
)

// NewSlackNotifier create new Thug bot
func NewSlackNotifier(slackToken string, channelID string) *SlackNotifier {
	slackNotifier := &SlackNotifier{
		SlackApi:  slack.New(slackToken),
		ChannelID: channelID,
	}
	go slackNotifier.run()
	return slackNotifier
}

func (t *SlackNotifier) messageHandler(ev *slack.MessageEvent) {
	if ev.Type == "message" &&
		(strings.HasPrefix(strings.ToLower(ev.Text), "hi bot") ||
			strings.HasPrefix(strings.ToLower(ev.Text), "hello bot")) {
		go t.helloWorld(ev)
		return
	}

	if ev.Type == "message" && strings.HasPrefix(strings.ToLower(ev.Text), "bot help") {
		go t.help(ev)
		return
	}
}

func (t *SlackNotifier) helloWorld(ev *slack.MessageEvent) (err error) {
	rand.Seed(time.Now().UnixNano())
	msg := greetingPrefixes[rand.Intn(len(greetingPrefixes))] + " <@" + ev.User + ">!"
	t.RTM.SendMessage(t.RTM.NewTypingMessage(ev.Channel))
	t.RTM.SendMessage(t.RTM.NewOutgoingMessage(msg, ev.Channel))
	return nil
}

func (t *SlackNotifier) help(ev *slack.MessageEvent) (err error) {
	t.RTM.SendMessage(t.RTM.NewOutgoingMessage(helpText, ev.Channel))
	return nil
}

func (t *SlackNotifier) run() {
	t.RTM = t.SlackApi.NewRTM()
	go t.RTM.ManageConnection()

	for msg := range t.RTM.IncomingEvents {
		switch ev := msg.Data.(type) {
		case *slack.ConnectedEvent:
			t.ID = ev.Info.User.ID
			log.Println(ev.Info.User.ID, "Connected")
		case *slack.MessageEvent:
			t.messageHandler(ev)
		case *slack.RTMError:
			log.Println(ev.Error())
		case *slack.InvalidAuthEvent:
			log.Println("Failed to Authenticate")
			return
		default:
		}
	}
}

func (s *SlackNotifier) Notify(results []Result) error {
	for _, result := range results {
		if result.Down {
			s.RTM.SendMessage(s.RTM.NewOutgoingMessage(result.Title+" is currently down", s.ChannelID))
		}
	}
	return nil
}
