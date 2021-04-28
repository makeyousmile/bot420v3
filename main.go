package main

import (
	"flag"
	"time"
)

type HydraShop struct {
	Category   string
	Title      string
	Text       string
	Market     string
	Price      string
	UpdateTime time.Time
}

type MessageToBot struct {
	id          int
	captcha     string
	captchaData string
	text        string
	stage       int
	hs          []HydraShop
	user        botUser
}
type MessageToWorker struct {
	id          int
	captcha     string
	captchaData string
	text        string
	mtype       int
	stage       int
	user        botUser
}
type Cfg struct {
	Accounts        []acc
	Proxy           string
	NumberOfWorkers int
	messageToBot    chan MessageToBot
	messageToWorker chan MessageToWorker
	TorProxy        string
	BotToken        string
	AdminChatId     int64
}

var (
	cfg = &Cfg{}
)

func init() {
	flag.StringVar(&cfg.TorProxy, "tor", "127.0.0.1:9150", "-tor ip:port")
	flag.StringVar(&cfg.BotToken, "token", "", "token")
	flag.Parse()
	cfg.AdminChatId = 150602226
	cfg.messageToBot = make(chan MessageToBot, 10)
	cfg.messageToWorker = make(chan MessageToWorker)
	cfg.Accounts = getAccs()
	cfg.Proxy = checkProxies(getProxies())
	cfg.NumberOfWorkers = len(getAccs())

}
func main() {

	go StartBot(cfg.messageToBot, cfg.messageToWorker)
	go StartCollyWorkers(cfg.messageToBot, cfg.messageToWorker)

	time.Sleep(24 * time.Hour)
}
