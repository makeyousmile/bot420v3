package main

import (
	"github.com/PuerkitoBio/goquery"
	"log"
	"strings"
	"time"
)

func parse(page string) ([]HydraShop, string) {
	var hydraShops []HydraShop
	selection, err := goquery.NewDocumentFromReader(strings.NewReader(page))
	if err != nil {
		log.Print(err)
	}
	cityName := ""
	selection.Find("select").Each(func(i int, selection *goquery.Selection) {
		value, exist := selection.Attr("name")
		if exist {
			if value == "region_id" {
				selection.Find("option").Each(func(i int, selection *goquery.Selection) {

					_, exist := selection.Attr("selected")
					if exist {
						cityName = strings.TrimSpace(selection.Text())
					}
				})
			}
		}
	})

	hs := HydraShop{}
	hs.Category = strings.TrimSpace(selection.Find("div.selected_category").Text())
	selection.Find("div.desc").Each(func(i int, selection *goquery.Selection) {

		hs.Title = strings.TrimSpace(selection.Find("div.title").Text())
		hs.Text = strings.TrimSpace(selection.Find("div.text").Text())
		hs.Market = strings.TrimSpace(selection.Find("div.market").Text())
		hs.Price = strings.TrimSpace(selection.Find("span.price").Text())
		hs.UpdateTime = time.Now()
		hydraShops = append(hydraShops, hs)
		selection.Find("div.slide").Each(func(n int, selection *goquery.Selection) {
			subhs := HydraShop{}
			subhs.Title = strings.TrimSpace(selection.Find("div.slide_title").Text())
			subhs.Price = strings.TrimSpace(selection.Find("span.slide_price").Text())
			subhs.Market = hs.Market
			subhs.UpdateTime = time.Now()
			hydraShops = append(hydraShops, subhs)
		})
	})

	return hydraShops, cityName
}
