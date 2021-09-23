package proxy

import (
	"bufio"
	"fmt"
	"log"
	"net/url"
	"os"
)

func ParseFromFile(path string) ([]*url.URL, error) {
	res := []*url.URL{}

	file, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
		return res, err
	}

	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		proxyURL, _ := url.Parse(fmt.Sprintf("http://%s", scanner.Text()))

		res = append(res, proxyURL)
	}

	if err = scanner.Err(); err != nil {
		return res, err
	}

	return res, nil
}
