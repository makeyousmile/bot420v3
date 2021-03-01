package main

import (
	"bytes"
	"github.com/PuerkitoBio/goquery"
	"github.com/avast/retry-go"
	"github.com/gocolly/colly"
	"github.com/gocolly/colly/proxy"
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
	userID       *int64
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
	c.UserAgent = "User Agent 3.57 11.86 Mozilla/5.0 (Windows NT 6.1; rv:60.0) Gecko/20100101 Firefox/60.0"
	//storage := &mongo.Storage{
	//	Database: "colly",
	//	URI:      "mongodb://localhost:27017",
	//}
	//if err := c.SetStorage(storage); err != nil {
	//	panic(err)
	//}
	c.ID = s.id
	c.WithTransport(&http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   60 * time.Second,
			KeepAlive: 60 * time.Second,
			//		DualStack: true,
		}).DialContext,
		MaxIdleConns:          0,
		IdleConnTimeout:       0,
		TLSHandshakeTimeout:   0,
		ExpectContinueTimeout: 0,
	})

	// Rotate two socks5 proxies
	//rp, err := proxy.RoundRobinProxySwitcher("socks5://165.232.72.180:9150")
	rp, err := proxy.RoundRobinProxySwitcher("socks5://127.0.0.1:9150")
	if err != nil {
		log.Fatal(err)
	}
	c.SetProxyFunc(rp)

	c.OnRequest(func(r *colly.Request) {
		//	fmt.Println("Visiting", r.URL)
	})

	c.OnResponse(func(r *colly.Response) {
		log.Print(*s.userID)
		//log.Print(s.CurrentStage)
		log.Printf("%s\n", bytes.Replace(r.Body, []byte("\n"), nil, -1))
		//log.Print(string(r.Body)[:])
		doc, err := goquery.NewDocumentFromReader(bytes.NewReader(r.Body))
		if err != nil {
			log.Print(err)
		}

		if strings.Contains(string(r.Body), "Забыли пароль?") {
			s.CurrentStage = 1
		}
		if s.CurrentStage == 3 {
			hydraShops, city := parse(string(r.Body))
			if len(hydraShops) > 0 {

				if city != "" {
					//	log.Print(r.Headers.Get("region_id") + "===================================================")
					city = TrimCollName(city)
					//WriteToDb(cityValues+":"+hydraShops[0].Category, hydraShops)
					msg := MessageToBot{
						id:    int(c.ID),
						stage: s.CurrentStage,
						hs:    hydraShops,
						user:  botUser{id: *s.userID},
					}
					log.Print(hydraShops)
					messageToBot <- msg

				}

				//log.Print(hydraShops[0].Market)
			}
			//c.Visit(link.getJob())
			return
		}
		Login := strings.Contains(string(r.Body), "Мои заказы")
		title := doc.Find("title").Text()
		if s.CurrentStage == 0 || title == "Вы не робот?" {
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
			err := c.Visit(cfg.Proxy + "login")
			if err != nil {
				log.Print(err)
			}

		} else if Login {
			s.CurrentStage = 3
			//log.Print("time to scrap!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!")
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

func StartCollyWorkers(messageToBot chan MessageToBot, messageToWorker chan MessageToWorker) {

	var scrapers []Scraper
	//	links := NewLinks()

	for i, account := range cfg.Accounts {
		scraper := Scraper{
			id:           uint32(i),
			CurrentStage: 0,
			Login:        account.Login,
			Pass:         account.Pass,
			Job:          CurrentValues{},
			captchaData:  "",
			userID:       new(int64),
		}
		scraper.collector = scraper.StartCollyWorker(messageToBot, messageToWorker)
		log.Print("scraper")
		log.Print(scraper.id)
		log.Print(scraper.userID)

		scrapers = append(scrapers, scraper)

	}
	for _, scraper := range scrapers {
		err := scraper.collector.Visit(cfg.Proxy)
		if err != nil {
			log.Print(err)
		}
		log.Print(scraper.id)
	}

	go func(s *[]Scraper) {
		for msg := range messageToWorker {

			if msg.mtype == 0 {
				if msg.stage == 0 {

					err := scrapers[msg.id].collector.Post(cfg.Proxy+"gate", map[string]string{
						"captcha":     msg.captcha,
						"captchaData": msg.captchaData,
					})
					if err != nil {
						log.Print(err)
					}

				} else if msg.stage == 2 {
					scrapers[msg.id].CurrentStage = msg.stage
					err := scrapers[msg.id].collector.Post(cfg.Proxy+"login", map[string]string{
						"captcha":     msg.captcha,
						"captchaData": msg.captchaData,
						"login":       scrapers[msg.id].Login,
						"password":    scrapers[msg.id].Pass,
					})
					if err != nil {
						log.Print(err)
					}

				} else if msg.stage == 3 {
					msgToBot := MessageToBot{
						id:          msg.id,
						captcha:     "",
						captchaData: "",
						text:        "time to scrap!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!",
						stage:       0,
					}
					messageToBot <- msgToBot

				}
			} else {
				if msg.mtype == 1 {
					*scrapers[0].userID = msg.user.id
					scrapers[0].CurrentStage = 10
					log.Print(scrapers[0].userID)
					log.Print(*scrapers[0].userID)
					job := cfg.Proxy + "catalog/" + msg.user.catValues + "?query=&region_id=" + msg.user.cityValues + "&subregion_id=0&price%5Bmin%5D=&price%5Bmax%5D=&unit=g&weight%5Bmin%5D=&weight%5Bmax%5D=&type=momental"

					err := retry.Do(
						func() error {
							err := scrapers[0].collector.Visit(job)
							if err != nil {
								return err
							}

							return nil
						},
					)
					if err != nil {
						msgToBot := MessageToBot{
							id:          msg.id,
							captcha:     "",
							captchaData: "",
							text:        err.Error(),
							stage:       0,
						}
						messageToBot <- msgToBot
					}
					log.Print(err)
					log.Print(job)
				}
			}

		}
	}(&scrapers)
}

func TrimCollName(collName string) string {
	name := strings.Split(collName, " ")
	return name[0]
}
