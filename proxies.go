package main

import (
	"bufio"
	"bytes"
	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly"
	"github.com/gocolly/colly/proxy"
	"log"
	"net"
	"net/http"
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
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		proxy := scanner.Text()
		proxies = append(proxies, proxy)
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
	sort.Strings(proxies)
	return proxies
}
func checkProxies(proxies []string) string {

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
		}

	}
	log.Print(workProxies)
	sort.Sort(workProxies)
	log.Print(workProxies)
	return workProxies[0].Addr
}

func startColly() *colly.Collector {
	c := colly.NewCollector(colly.AllowURLRevisit())
	c.UserAgent = "User Agent 3.57 11.86 Mozilla/5.0 (Windows NT 6.1; rv:60.0) Gecko/20100101 Firefox/60.0"
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
	return c
}
func checkProxy(addr string) (bool, time.Duration) {
	var check bool
	c := startColly()

	rp, err := proxy.RoundRobinProxySwitcher("socks5://" + cfg.TorProxy)
	if err != nil {
		log.Fatal(err)
	}
	c.SetProxyFunc(rp)

	c.OnRequest(func(r *colly.Request) {
		//	fmt.Println("Visiting", r.URL)
	})

	c.OnResponse(func(r *colly.Response) {
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
