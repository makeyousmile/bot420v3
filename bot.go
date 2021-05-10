package main

import (
	tgbotapi "github.com/Syfaro/telegram-bot-api"
	"log"
	"strconv"
	"time"
)

func StartBot(messagesToBot chan MessageToBot, messagesToWorker chan MessageToWorker) {

	checkWithInterval(messagesToWorker, 5)

	links := NewLinks()
	users := make(map[int64]botUser)

	//start new Telegram Bot with API token from Cfg struct var
	bot, err := tgbotapi.NewBotAPI(cfg.BotToken)
	if err != nil {
		log.Fatal(err)
	}
	bot.Debug = true
	log.Printf("Authorized on account %s", bot.Self.UserName)
	ucfg := tgbotapi.NewUpdate(0)
	ucfg.Timeout = 60

	cash := map[int]MessageToBot{}

	go func() {
		for msg := range messagesToBot {

			cash[msg.id] = msg

			if msg.text != "" {
				text := tgbotapi.NewMessage(cfg.AdminChatId, msg.text+" id="+strconv.FormatUint(uint64(msg.id), 10))
				_, err := bot.Send(text)
				if err != nil {
					log.Print(err)
				}
				continue
			}
			if msg.stage == 0 {

				photo := tgbotapi.NewPhotoUpload(cfg.AdminChatId, strconv.FormatUint(uint64(msg.id), 10)+".jpeg")
				photo.Caption = strconv.FormatUint(uint64(msg.id), 10)
				_, err := bot.Send(photo)
				if err != nil {
					log.Print(err)
				}
			}
			if msg.stage == 2 {

				//msgtext := tgbotapi.NewMessage(150602226, msg.captcha)
				photo := tgbotapi.NewPhotoUpload(cfg.AdminChatId, strconv.FormatUint(uint64(msg.id), 10)+".jpeg")
				photo.Caption = strconv.FormatUint(uint64(msg.id), 10)
				_, err := bot.Send(photo)
				if err != nil {
					log.Print(err)
				}
			}
			if msg.stage == 3 {
				log.Print("msg.stage == 3")
				answer := marketView(msg.hs)
				//for i, market := range msg.hs {
				//	if market.Price != "" {
				//		m := strconv.Itoa(i+1) + ". <b>" + market.Title + "</b>\n " + market.Price + "\n\n"
				//		answer += m
				//	}
				//}
				msg := tgbotapi.NewMessage(msg.user.id, answer)
				msg.ParseMode = "HTML"
				_, err := bot.Send(msg)
				if err != nil {
					log.Print(err)
				}
			}
			if msg.user.id != 0 {
				log.Print(msg.user.id != 0)
			}

		}
	}()

	updates, err := bot.GetUpdatesChan(ucfg)
	for update := range updates {

		if update.Message != nil {
			id := update.Message.Chat.ID
			user := botUser{}
			users[id] = user

			typing := tgbotapi.NewChatAction(update.Message.Chat.ID, "typing")
			_, err := bot.Send(typing)
			if err != nil {
				log.Print(err)
			}

			for i := 0; i < cfg.NumberOfWorkers; i++ {
				if update.Message.Command() == strconv.Itoa(i) {

					msg := tgbotapi.NewMessage(update.Message.Chat.ID, strconv.Itoa(i))
					msg.ReplyToMessageID = update.Message.MessageID
					_, err := bot.Send(msg)
					if err != nil {
						log.Print(err)
					}

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
				_, err = bot.Send(msg)
				if err != nil {
					return
				}
			}

			if update.Message.Command() == "check" {
				text := tgbotapi.NewMessage(update.Message.Chat.ID, "Проверка работоспособности зеркал (1-2 минуты) ... ")
				bot.Send(text)
				answer := ""
				mirrors := checkProxies(getProxies())

				for i, mirror := range mirrors {
					answer += "\n" + strconv.Itoa(i) + ". "
					if mirror.ResTime < cfg.ResponseTimeLimit {
						answer += mirror.Addr + "Время отклика: " + mirror.ResTime.String()
					} else {
						answer += "Зеркало: " + mirror.Addr + " недоступно!"
					}
				}
				text = tgbotapi.NewMessage(update.Message.Chat.ID, answer)
				bot.Send(text)

			}
			if update.Message.Command() == "go" {

				text := tgbotapi.NewMessage(cfg.AdminChatId, strconv.FormatUint(uint64(update.Message.Chat.ID), 10)+" "+update.Message.Chat.UserName+" "+update.Message.Chat.FirstName+" "+update.Message.Chat.LastName)
				_, err := bot.Send(text)
				if err != nil {
					log.Print(err)
				}

				var numericKeyboard tgbotapi.InlineKeyboardMarkup
				for i, city := range cityNames {
					row := tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData(city, cityValues[i]))
					numericKeyboard.InlineKeyboard = append(numericKeyboard.InlineKeyboard, row)
				}

				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Выберите город")
				msg.ReplyMarkup = numericKeyboard
				_, err = bot.Send(msg)
				if err != nil {
					log.Print(err)
				}
				log.Print("hey")
			}
		}
		if update.CallbackQuery != nil {
			id := update.CallbackQuery.Message.Chat.ID
			_, err := bot.AnswerCallbackQuery(tgbotapi.NewCallback(update.CallbackQuery.ID, update.CallbackQuery.Data))
			if err != nil {
				log.Print(err)
			}
			data := update.CallbackQuery.Data

			if users[id].cityValues == "" {
				users[id] = botUser{cityValues: data}

			} else {
				users[id] = botUser{
					cityValues: users[id].cityValues,
					catValues:  data,
					id:         id,
				}
				//editedMsg := tgbotapi.NewEditMessageText(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID, users[id].cityValues+users[id].catValues)
				editedMsg := tgbotapi.NewDeleteMessage(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID)
				_, err := bot.Send(editedMsg)
				if err != nil {
					log.Print(err)
				}
				mess := MessageToWorker{
					mtype: 1,
					user:  users[id],
					stage: 10,
				}
				messagesToWorker <- mess
			}
			log.Print(users)

			if users[id].catValues == "" {
				var CatKeyboard tgbotapi.InlineKeyboardMarkup

				for i, cat := range catNames {
					row := tgbotapi.NewInlineKeyboardRow(
						tgbotapi.NewInlineKeyboardButtonData(cat, catValues[i]),
					)
					CatKeyboard.InlineKeyboard = append(CatKeyboard.InlineKeyboard, row)
				}
				//edit top text
				editedMsg := tgbotapi.NewEditMessageText(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID, "Выберите категорию")
				_, err := bot.Send(editedMsg)
				if err != nil {
					log.Print(err)
				}
				//edit body
				editedMsg2 := tgbotapi.NewEditMessageReplyMarkup(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID, CatKeyboard)
				_, err = bot.Send(editedMsg2)
				if err != nil {
					log.Print(err)
				}
			}

		}

	}
}

func marketView(markets []HydraShop) string {
	if len(markets) == 0 {
		return "<b> Пустая позиция </b>"
	}
	var view string
	for i := 0; i < len(markets); i++ {
		market := []rune(markets[i].Market)
		if len(market) > 39 {
			market = market[:38]
		}
		view += strconv.Itoa(i+1) + ". " + markets[i].Title + "\n  " + "<b>" + markets[i].Price + "</b>" + "\n <code>" + string(market) + "</code>\n\n"

	}
	return view
}

func checkWithInterval(bot chan MessageToWorker, interval int) {
	go func() {
		for {
			log.Println("---------------------------------checkInterval------------------------------")
			user := botUser{
				cityValues: "410",
				catValues:  "3",
				id:         cfg.AdminChatId,
			}
			m := MessageToWorker{
				mtype: 1,
				stage: 10,
				user:  user,
			}
			bot <- m
			time.Sleep(time.Minute * time.Duration(interval))
		}
	}()

}
