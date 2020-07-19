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

var NumberOfWorkers = 3

var hydraProxy = getProxies()[3]

func main() {

	accs := getAccs()

	messageToBot := make(chan MessageToBot)
	messageToWorker := make(chan MessageToWorker)

	go StartBot(messageToBot, messageToWorker)
	go StartCollyWorkers(messageToBot, messageToWorker, accs)

	time.Sleep(10000 * time.Second)
}
