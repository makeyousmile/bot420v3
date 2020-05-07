package main

import (
	"bytes"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly"
	"github.com/gocolly/colly/proxy"
	"github.com/gocolly/colly/queue"
	"log"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type CurrentValues struct {
	cityValue string
	catValue  string
}
type Scraper struct {
	id           uint32
	collector    *colly.Collector
	CurrentStage int
	Login        string
	Pass         string
	Job          CurrentValues
	captcha      string
	captchaData  string
}

func (s *Scraper) StartCollyWorker(messageToBot chan MessageToBot, messageToWorker chan MessageToWorker) *colly.Collector {
	//link := NewLinks()
	c := colly.NewCollector(colly.AllowURLRevisit())
	c.ID = s.id
	c.WithTransport(&http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   30000 * time.Second,
			KeepAlive: 30000 * time.Second,
			DualStack: true,
		}).DialContext,
		MaxIdleConns:          0,
		IdleConnTimeout:       0,
		TLSHandshakeTimeout:   0,
		ExpectContinueTimeout: 0,
	})
	// Rotate two socks5 proxies
	rp, err := proxy.RoundRobinProxySwitcher("socks5://127.0.0.1:9150")
	if err != nil {
		log.Fatal(err)
	}
	c.SetProxyFunc(rp)

	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL.String())
	})

	c.OnResponse(func(r *colly.Response) {

		//	log.Printf("%s\n", bytes.Replace(r.Body, []byte("\n"), nil, -1))
		//log.Print(string(r.Body)[:])
		doc, err := goquery.NewDocumentFromReader(bytes.NewReader(r.Body))
		if err != nil {
			log.Print(err)
		}
		title := doc.Find("title").Text()
		Login := strings.Contains(string(r.Body), s.Login)
		if s.CurrentStage == 3 {
			hs, city := parse(string(r.Body))
			if len(hs) > 0 {
				//
				if city != "" {
					WriteToDb(city, hs)
				}

				//log.Print(hs[0].Market)
			}
			//c.Visit(link.getJob())
			return
		}
		if s.CurrentStage == 0 && title == "Вы не робот?" {
			file, exist := doc.Find("img").Attr("src")
			if exist {
				base64toJpg(file, strconv.FormatUint(uint64(c.ID), 10)+".jpeg")
			}

			doc.Find("input").Each(func(i int, selection *goquery.Selection) {
				tag, exist := selection.Attr("name")
				if exist {
					if tag == "captchaData" {
						value, exist := selection.Attr("value")
						if exist {
							s.captchaData = value
							msg := MessageToBot{
								id:          int(c.ID),
								captchaData: value,
								stage:       s.CurrentStage,
							}
							messageToBot <- msg
						}
					}
				}

			})
		} else if s.CurrentStage == 0 && title == "HYDRA" {
			s.CurrentStage = 1
			log.Print("stage = 1")
			log.Print(s.CurrentStage)
			log.Print("visit login stage ")
			c.Visit("http://hydraruzxpnew4af.onion/login")

		} else if Login {
			s.CurrentStage = 3
			//log.Print("time to scrab!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!")
			go func() {
				msg := MessageToWorker{
					id:          int(s.id),
					captcha:     "",
					captchaData: "",
					text:        "",
					mtype:       0,
					stage:       s.CurrentStage,
				}
				messageToWorker <- msg
			}()

		} else if s.CurrentStage == 2 && title == "HYDRA" {
			log.Print("stage = 2 2")
			file, exist := doc.Find("img").Attr("src")
			if exist {
				base64toJpg(file, strconv.FormatUint(uint64(c.ID), 10)+".jpeg")
			}

			doc.Find("input").Each(func(i int, selection *goquery.Selection) {
				tag, exist := selection.Attr("name")
				if exist {
					if tag == "captchaData" {
						value, exist := selection.Attr("value")
						if exist {
							msg := MessageToBot{
								id:          int(c.ID),
								captchaData: value,
								stage:       s.CurrentStage,
							}
							messageToBot <- msg
						}
					}
				}

			})

		} else if s.CurrentStage == 1 && title == "HYDRA" {
			log.Print("stage = 2")
			s.CurrentStage = 2
			file, exist := doc.Find("img").Attr("src")
			if exist {
				base64toJpg(file, strconv.FormatUint(uint64(c.ID), 10)+".jpeg")
			}

			doc.Find("input").Each(func(i int, selection *goquery.Selection) {
				tag, exist := selection.Attr("name")
				if exist {
					if tag == "captchaData" {
						value, exist := selection.Attr("value")
						if exist {
							msg := MessageToBot{
								id:          int(c.ID),
								captchaData: value,
								stage:       s.CurrentStage,
							}

							messageToBot <- msg
						}
					}
				}

			})

		}

		//log.Print(doc.Find("img").Attr("src"))

	})

	log.Print("start colly")

	return c
}

func StartCollyWorkers(messageToBot chan MessageToBot, messageToWorker chan MessageToWorker, accounts []acc) {

	scrapers := []Scraper{}
	links := NewLinks()
	q, _ := queue.New(
		2, // Number of consumer threads
		&queue.InMemoryQueueStorage{MaxSize: 10000}, // Use default queue storage
	)

	for i, account := range accounts {
		scraper := Scraper{
			id:           uint32(i),
			CurrentStage: 0,
			Login:        account.Login,
			Pass:         account.Pass,
			Job:          CurrentValues{},
			captchaData:  "",
		}
		scraper.collector = scraper.StartCollyWorker(messageToBot, messageToWorker)

		scrapers = append(scrapers, scraper)

	}
	for _, scraper := range scrapers {
		scraper.collector.Visit("http://hydraruzxpnew4af.onion")
		log.Print(scraper.id)
	}

	go func(s []Scraper) {
		for msg := range messageToWorker {

			if msg.mtype == 0 {
				if msg.stage == 0 {

					err := scrapers[msg.id].collector.Post("http://hydraruzxpnew4af.onion/gate", map[string]string{
						"captcha":     msg.captcha,
						"captchaData": msg.captchaData,
					})
					if err != nil {
						log.Print(err)
					}

				} else if msg.stage == 2 {
					scrapers[msg.id].CurrentStage = msg.stage
					err := scrapers[msg.id].collector.Post("http://hydraruzxpnew4af.onion/login", map[string]string{
						"captcha":     msg.captcha,
						"captchaData": msg.captchaData,
						"login":       scrapers[msg.id].Login,
						"password":    scrapers[msg.id].Pass,
					})
					if err != nil {
						log.Print(err)
					}

				} else if msg.stage == 3 {
					scrapers[msg.id].CurrentStage = msg.stage
					log.Print("-----------------------------------------start jobs-----------------------------------------start jobs")
					//scrapers[msg.id].collector.Visit(links.getJob())
					for _, job := range links.getJobs() {
						q.AddURL(job)
					}
					err := q.Run(scrapers[msg.id].collector)
					if err != nil {
						log.Print(err)
					}
				}
			}
		}
	}(scrapers)
}
