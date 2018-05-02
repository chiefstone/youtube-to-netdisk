package broadcaster

import (
	"time"
	tg "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/yanzay/log"
	"github.com/cool2645/youtube-to-netdisk/model"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"sync"
	"github.com/rikakomoe/ritorudemonriri/ririsdk"
	"encoding/json"
)

var subscribedChats = make(map[int64]int64)
var mux sync.RWMutex
var ch = make(chan string)

func ServeTelegram(db *gorm.DB, addr string, key string) {
	log.Infof("Reading subscribed chats from database %s", time.Now())
	subscribers, err := model.ListSubscribers(db)
	if err != nil {
		log.Fatal(err)
	}
	for _, v := range subscribers {
		subscribedChats[v] = v
	}
	log.Warningf("%v %s", subscribedChats, time.Now())
	log.Infof("Started serve telegram %s", time.Now())
	ririsdk.Init(addr, key, true)
	go pushMessage(ch)
	messages, _ := ririsdk.GetUpdatesChan()
	for message := range messages {
		b, err := json.Marshal(message)
		if err != nil {
			log.Error(err)
			continue
		}
		log.Info(string(b))
		if message.Direction != ririsdk.IN {
			continue
		}
		switch message.Messenger {
		case ririsdk.Telegram:
			m := message.TelegramMessage
			if m.IsCommand() {
				switch m.Command() {
				case "carrier_subscribe":
					ReplyMessage(start(db, m), m.Chat.ID)
				case "carrier_unsubscribe":
					ReplyMessage(stop(db, m), m.Chat.ID)
				}
			}
		}
	}
}

func replyMessage(text string, parseMode string, reqChatID int64) {
	msg := tg.NewMessage(reqChatID, text)
	msg.ParseMode = parseMode
	ririsdk.PushMessage(ririsdk.Message{
		Direction:             ririsdk.OUT,
		Messenger:             ririsdk.Telegram,
		TelegramMessageConfig: &msg,
	})
}

func pushMessage(c chan string) {
	var m string
	for {
		m = <-c
		mux.RLock()
		for _, v := range subscribedChats {
			ReplyMessage(m, v)
		}
		mux.RUnlock()
	}
}

func start(db *gorm.DB, m *tg.Message) string {
	mux.Lock()
	defer mux.Unlock()
	subscribedChats[m.Chat.ID] = m.Chat.ID
	_, err := model.SaveSubscriber(db, m.Chat.ID)
	if err != nil {
		log.Fatal(err)
	}
	return "You have set up subscription of yt2nd for this chat, pwp"
}

func stop(db *gorm.DB, m *tg.Message) string {
	mux.Lock()
	defer mux.Unlock()
	delete(subscribedChats, m.Chat.ID)
	err := model.RemoveSubscriber(db, m.Chat.ID)
	if err != nil {
		log.Fatal(err)
	}
	return "Your subscription of yt2nd is suspended, qaq"
}

func ReplyMessage(text string, reqChatID int64) {
	replyMessage(text, "Markdown", reqChatID)
}

func Broadcast(msg string) {
	ch <- msg
	return
}
