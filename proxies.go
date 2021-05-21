package main

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly"
	"github.com/gocolly/colly/proxy"
	"log"
	"os"
	"sort"
	"time"
)

type Mirror struct {
	Addr    string
	ResTime time.Duration
}

type Mirrors []Mirror

func (m Mirrors) Len() int {
	return len(m)
}

func (m Mirrors) Swap(i, j int) {
	m[i], m[j] = m[j], m[i]
}
func (m Mirrors) Less(i, j int) bool {
	return m[i].ResTime < m[j].ResTime
}

func getProxies() []string {
	var proxies []string
	file, err := os.Open("proxies.txt")
	if err != nil {
		log.Fatal(err)
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			log.Println(err)
		}
	}(file)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		mirror := scanner.Text()
		proxies = append(proxies, mirror)
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
	return proxies
}
func checkProxies(proxies []string) Mirrors {

	var workProxies Mirrors

	// Rotate two socks5 proxies
	for _, addr := range proxies {
		check, responseTime := checkProxy(addr)
		if check {
			msg := MessageToBot{
				text: addr + "" + responseTime.String(),
			}
			cfg.messageToBot <- msg
			mirror := Mirror{
				Addr:    addr,
				ResTime: responseTime,
			}
			workProxies = append(workProxies, mirror)
		} else {
			mirror := Mirror{
				Addr:    addr,
				ResTime: time.Second * 999,
			}
			workProxies = append(workProxies, mirror)
		}

	}
	log.Print(workProxies)
	sort.Sort(workProxies)
	log.Print(workProxies)
	if len(workProxies) == 0 {
		log.Fatal("Zero working mirrors found. Exiting ...")
	}
	return workProxies
}

func checkProxy(addr string) (bool, time.Duration) {
	var check bool
	c := NewColly()
	rp, err := proxy.RoundRobinProxySwitcher("socks5://" + cfg.TorProxy)
	if err != nil {
		log.Fatal(err)
	}
	c.SetProxyFunc(rp)

	c.OnError(func(r *colly.Response, err error) {
		log.Println(r.Request.URL)
		log.Print(err)
	})
	c.OnRequest(func(r *colly.Request) {
		// EOF fix from so
		r.Headers.Set("Accept-Encoding", "gzip")

		fmt.Println("Visiting", r.URL)
	})

	c.OnResponse(func(r *colly.Response) {
		log.Print(r.Request.URL)
		doc, err := goquery.NewDocumentFromReader(bytes.NewReader(r.Body))
		if err != nil {
			log.Print(err)
		}
		title := doc.Find("title").Text()
		if title == "Вы не робот?" {
			log.Print("send: true")
			check = true
		} else {
			log.Print("send: false")
			check = false
		}
	})
	log.Print("check1")
	c.Async = false
	t1 := time.Now()
	err = c.Visit(addr)
	if err != nil {
		log.Print(err)
		go func() {
			msg := MessageToBot{
				text: err.Error(),
			}
			cfg.messageToBot <- msg
		}()
	}

	c.Wait()
	t2 := time.Now().Sub(t1)

	return check, t2
}
