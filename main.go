package main

import (
	"time"
)

type MessageToBot struct {
	id          int
	captcha     string
	captchaData string
	text        string
	stage       int
}
type MessageToWorker struct {
	id          int
	captcha     string
	captchaData string
	text        string
	mtype       int
	stage       int
}

var (
	accs            = getAccs()
	NumberOfWorkers = len(accs)
	hydraProxy      = getProxies()[2]
)

func main() {

	messageToBot := make(chan MessageToBot)
	messageToWorker := make(chan MessageToWorker)

	go StartBot(messageToBot, messageToWorker)
	go StartCollyWorkers(messageToBot, messageToWorker, accs)

	time.Sleep(24 * time.Hour)
}
