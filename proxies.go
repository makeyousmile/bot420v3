package main

import (
	"bufio"
	"log"
	"os"
)

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
	return proxies
}
