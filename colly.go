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

	c := NewColly()
	c.ID = s.id
	c.OnRequest(func(r *colly.Request) {
		log.Print(r.URL.String())
		// EOF fix from so
		r.Headers.Set("Accept-Encoding", "gzip")

	})

	c.OnResponse(func(r *colly.Response) {
		temphs := r.Ctx.GetAny("hs")
		// если temphs не пуст ->  парсим страницу с позицией
		if temphs != nil {
			hs := temphs.(HydraShop)
			log.Print("888888888888888888888888888888888888888888888888888888888888888888888888888888888888888888888888888888888888888888888888888888888888")
			log.Print(cfg.Proxy + hs.Link)

			doc, err := goquery.NewDocumentFromReader(bytes.NewReader(r.Body))
			if err != nil {
				log.Print(err)
			}
			city := r.Ctx.Get("city")

			doc.Find("li.momental-region-" + city).Each(func(i int, selection *goquery.Selection) {
				log.Print("ssssssssssssssssssssssssssssssssssssssssssssssssssssssssssssssssssssssssssssssssssssssssssssssssssssssssssssssssssssssssssssssssssssssssssssssssssssssssss")
				log.Print(selection.Text())
			})

		} else {

			//log.Print(string(r.Body)[:])
			doc, err := goquery.NewDocumentFromReader(bytes.NewReader(r.Body))
			if err != nil {
				log.Print(err)
			}

			if strings.Contains(string(r.Body), "Забыли пароль?") {
				s.CurrentStage = 1
			}
			if s.CurrentStage == 3 {
				city := r.Ctx.Get("city")
				hydraShops := parse(string(r.Body))
				for _, hydraShop := range hydraShops {
					log.Print(hydraShop)
					ctx := colly.NewContext()
					ctx.Put("hs", hydraShop)
					ctx.Put("city", city)
					err := s.collector.Request("GET", cfg.Proxy+hydraShop.Link, nil, ctx, nil)
					if err != nil {
						log.Print("get Position Page Error")
					}
				}

				msg := MessageToBot{
					id:    int(c.ID),
					stage: s.CurrentStage,
					hs:    hydraShops,
					user:  botUser{id: *s.userID},
				}
				log.Print(hydraShops)
				messageToBot <- msg

				return
			}
			Login := strings.Contains(string(r.Body), "Мои заказы")
			title := doc.Find("title").Text()
			switch s.CurrentStage {

			}
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
				switch msg.stage {
				case 0:
					err := scrapers[msg.id].collector.Post(cfg.Proxy+"gate", map[string]string{
						"captcha":     msg.captcha,
						"captchaData": msg.captchaData,
					})
					if err != nil {
						log.Print(err)
					}
				case 2:
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
				case 3:
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

					job := cfg.Proxy + "catalog/" + msg.user.catValues + "?query=&region_id=" + msg.user.cityValues + "&subregion_id=0&price%5Bmin%5D=&price%5Bmax%5D=&unit=g&weight%5Bmin%5D=&weight%5Bmax%5D=&type=momental"
					retry.DefaultAttempts = 3
					ctx := colly.NewContext()
					ctx.Put("city", msg.user.cityValues)

					err := retry.Do(
						func() error {
							//err := scrapers[0].collector.Visit(job)
							err := scrapers[0].collector.Request("GET", job, nil, ctx, nil)
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
					//переодеческий запрос
					go func() {
						for {
							time.Sleep(time.Minute * 5)
							err := scrapers[0].collector.Request("GET", cfg.Proxy, nil, ctx, nil)
							if err != nil {
								MessageToAdmin(messageToBot, err.Error())
								log.Print(err)
							}

						}
					}()
					log.Print(err)
					log.Print(job)
				}
			}

		}
	}(&scrapers)
}

func NewColly() *colly.Collector {
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
	//c.ID = s.id
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
	c.SetRequestTimeout(60 * time.Second)

	// Rotate two socks5 proxies
	//rp, err := proxy.RoundRobinProxySwitcher("socks5://165.232.72.180:9150")
	proxyString := "socks5://" + cfg.TorProxy
	rp, err := proxy.RoundRobinProxySwitcher(proxyString)
	if err != nil {
		log.Fatal(err)
	}
	c.SetProxyFunc(rp)
	return c
}
func MessageToAdmin(m chan MessageToBot, s string) {
	msgToBot := MessageToBot{
		text: s,
	}
	m <- msgToBot
}
