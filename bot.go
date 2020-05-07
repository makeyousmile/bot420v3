package main

import (
	"encoding/json"
	tgbotapi "github.com/Syfaro/telegram-bot-api"
	"log"
	"os"
	"strconv"
)

type Config struct {
	TelegramBotToken string
}
type postForm struct {
	CapchaData string
	Capcha     string
	Login      string
	Password   string
}

func StartBot(messagesToBot chan MessageToBot, messagesToWorker chan MessageToWorker) {

	links := NewLinks()

	file, _ := os.Open("config.json")
	decoder := json.NewDecoder(file)
	cfg := Config{}
	err := decoder.Decode(&cfg)
	if err != nil {
		log.Fatal(err)
	}
	//start new Telegram Bot with API token from Config struct var
	bot, err := tgbotapi.NewBotAPI(cfg.TelegramBotToken)
	if err != nil {
		log.Fatal(err)
	}

	cash := map[int]MessageToBot{}

	go func() {
		for msg := range messagesToBot {

			cash[msg.id] = msg

			if msg.text != "" {
				text := tgbotapi.NewMessage(150602226, msg.text+" id="+strconv.FormatUint(uint64(msg.id), 10))
				bot.Send(text)
				continue
			}
			if msg.stage == 0 {

				photo := tgbotapi.NewPhotoUpload(150602226, strconv.FormatUint(uint64(msg.id), 10)+".jpeg")
				photo.Caption = strconv.FormatUint(uint64(msg.id), 10)
				bot.Send(photo)
			}
			if msg.stage == 2 {

				//msgtext := tgbotapi.NewMessage(150602226, msg.captcha)
				photo := tgbotapi.NewPhotoUpload(150602226, strconv.FormatUint(uint64(msg.id), 10)+".jpeg")
				photo.Caption = strconv.FormatUint(uint64(msg.id), 10)
				bot.Send(photo)
			}

		}
	}()

	bot.Debug = false
	log.Printf("Authorized on account %s", bot.Self.UserName)
	ucfg := tgbotapi.NewUpdate(0)
	ucfg.Timeout = 60

	updates, err := bot.GetUpdatesChan(ucfg)
	for update := range updates {

		if update.Message == nil {
			continue
		}
		typing := tgbotapi.NewChatAction(update.Message.Chat.ID, "typing")
		bot.Send(typing)

		for i := 0; i < NumberOfWorkers; i++ {
			if update.Message.Command() == strconv.Itoa(i) {

				msg := tgbotapi.NewMessage(update.Message.Chat.ID, strconv.Itoa(i))
				msg.ReplyToMessageID = update.Message.MessageID
				bot.Send(msg)

				log.Print(cash[i].captchaData)
				log.Print(update.Message.CommandArguments())
				mess := MessageToWorker{
					id:          i,
					text:        "",
					captchaData: cash[i].captchaData,
					captcha:     update.Message.CommandArguments(),
					mtype:       0,
					stage:       cash[i].stage,
				}
				messagesToWorker <- mess
				//log.Print(workers[i].Login)
				//log.Print(workers[i].Pass)
				//err := workers[i].collector.Post("http://hydraruzxpnew4af.onion/gate", map[string]string{
				//	"captcha":     captcha,
				//	"captchaData": captchaData,
				//})
				if err != nil {
					log.Print(err)
				}

			}

		}

		if update.Message.Command() == "link" {
			//link := update.MessageToBot.CommandArguments()
			job := links.getJob()
			log.Print("visit: " + job)
			worker, err := strconv.Atoi(update.Message.CommandArguments())
			if err != nil {
				log.Print(err)
				continue
			}
			log.Print(worker)
			//workers[worker].collector.Visit(job)
			//workers[0].collector.Visit(job)
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "1")
			msg.ReplyToMessageID = update.Message.MessageID
			bot.Send(msg)
		}

		if update.Message.Command() == "test" {
			log.Print(update.Message.CommandArguments())
		}
	}
}
