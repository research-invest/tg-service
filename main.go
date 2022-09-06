package main

//GOOS=linux GOARCH=amd64 go build -o ./tg-service -a

import (
	"bytes"
	"encoding/json"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/sirupsen/logrus"
	"net/http"
	"os"
	"strconv"
	"strings"
)

const ChannelUrl = "/channel/"

type ReplyMarkup struct {
	KeyboardButtonRows [][]string `json:"keyboardButtonRows"`
}

type SendMessageChannel struct {
	ChatId      int64       `json:"chat_id"`
	Text        string      `json:"text"`
	ReplyMarkup ReplyMarkup `json:"replyMarkup"`
}

var ChannelBots = make(map[string]*tgbotapi.BotAPI, 5) // 5!!!!!

var log = logrus.New()

func main() {
	setLogParam()
	readConfig()
	runListenChannels()
	startApiListenAndServe()
}

func runListenChannels() {
	for _, channel := range appConfig.Channels {
		channel := channel

		bot, err := tgbotapi.NewBotAPI(channel.Token)
		if err != nil {
			log.Panic(err)
		}

		log.Printf("Authorized on account %s", bot.Self.UserName)

		bot.Debug = false // false
		ChannelBots[channel.UrlCode] = bot

		HttpHandlerFunc(ChannelUrl+channel.UrlCode, http.HandlerFunc(ChannelHandler))

		go func() {
			startListenerChannel(channel)
		}()
	}
}

func startApiListenAndServe() {

	port := appConfig.Port
	host := appConfig.Host

	fmt.Println("Start service on " + host + ":" + strconv.Itoa(port) + ".")

	err := http.ListenAndServe(host+":"+strconv.Itoa(port), nil)
	if err != nil {
		log.Fatalf("can't run service : %v", err)
	}
}

func setLogParam() {
	log.Out = os.Stdout

	file, err := os.OpenFile("./logs/logrus.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err == nil {
		log.Out = file
	} else {
		log.Info("Failed to log to file, using default stderr")
	}
}

func HttpHandlerFunc(pattern string, h http.Handler) {
	http.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				WriteHeader(w, fmt.Sprintf("%v - panic occurred:%v", pattern, err))
			}
		}()

		h.ServeHTTP(w, r)
	})
}

func WriteHeader(w http.ResponseWriter, response interface{}) {
	responseJson, err := json.Marshal(response)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		responseWrite(w, []byte(fmt.Sprintf("can't marshal json : %v", err)))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	responseWrite(w, responseJson)
}

func responseWrite(w http.ResponseWriter, data []byte) {
	_, err := w.Write(data)
	if err != nil {
		log.Warnf("can't write response : %v", err)
	}
}

func startListenerChannel(channel Channel) {
	bot := ChannelBots[channel.UrlCode]

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, _ := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil { // ignore any non-Message updates
			continue
		}

		postBody, _ := json.Marshal(map[string]interface{}{
			"command":      update.Message.Command(),
			"chat_id":      update.Message.Chat.ID,
			"first_name":   update.Message.Chat.FirstName,
			"last_name":    update.Message.Chat.LastName,
			"username":     update.Message.Chat.UserName,
			"text_message": update.Message.Text,
		})

		responseBody := bytes.NewBuffer(postBody)

		resp, err := http.Post(channel.UrlApi, "application/json", responseBody)
		if err != nil {
			log.Error("An Error Occurred %v", err)

			sendMessage := &SendMessageChannel{
				ChatId: update.Message.Chat.ID,
				Text:   "Error #21353",
			}

			err, _ = sendMessageInChannel(channel.UrlCode, sendMessage)
			if err != nil {
				log.Error("An Error sendMessageInChannel %v", err)
				continue
			}

			continue
		}

		defer resp.Body.Close()

		d := json.NewDecoder(resp.Body)
		sendMessage := &SendMessageChannel{}
		err = d.Decode(sendMessage)

		if err != nil {
			log.Error("An Error read from api response %v", err)
			continue
		}

		sendMessage.ChatId = update.Message.Chat.ID

		err, _ = sendMessageInChannel(channel.UrlCode, sendMessage)
		if err != nil {
			log.Error("An Error read from api response send %v", err)
			continue
		}
	}
}

func ChannelHandler(w http.ResponseWriter, r *http.Request) {
	code := strings.Replace(r.URL.Path, ChannelUrl, "", 1)
	switch r.Method {
	case "POST":
		d := json.NewDecoder(r.Body)
		sendMessage := &SendMessageChannel{}

		err := d.Decode(sendMessage)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			log.Error("ChannelHandler error Decode %v", err)
			return
		}

		err, _ = sendMessageInChannel(code, sendMessage)
		if err != nil {
			log.Error("ChannelHandler error send message %v", err)
			return
		}

		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "success")
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		fmt.Fprintf(w, "I can't do that.")
	}
}

func sendMessageInChannel(code string, sendMessage *SendMessageChannel) (error, bool) {
	bot := ChannelBots[code]
	msg := tgbotapi.NewMessage(sendMessage.ChatId, sendMessage.Text)
	msg.ParseMode = "MarkdownV2"
	//msg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)

	if sendMessage.ReplyMarkup.KeyboardButtonRows != nil {
		rows := make([][]tgbotapi.KeyboardButton, len(sendMessage.ReplyMarkup.KeyboardButtonRows))
		for rowIndex, rowButtons := range sendMessage.ReplyMarkup.KeyboardButtonRows {
			buttons := make([]tgbotapi.KeyboardButton, len(rowButtons))
			for btnIndex, btnTxt := range rowButtons {
				buttons[btnIndex] = tgbotapi.NewKeyboardButton(btnTxt)
			}
			rows[rowIndex] = tgbotapi.NewKeyboardButtonRow(buttons...)
		}

		msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(rows...)
	}

	if _, err := bot.Send(msg); err != nil {
		log.Error("sendMessageInChannel error send message %v", err)
		return err, false
	}

	return nil, true
}
