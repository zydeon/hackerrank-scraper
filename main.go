package main

import (
	"encoding/json"
	"fmt"
	"log"
	"regexp"

	"github.com/PuerkitoBio/goquery"
)

type Category struct {
	Name, Slug string
	Categories []Category `json:"children"`
}

type Challenge struct {
	Name, Slug string
	Category   Category
}

func main() {
	cs := readCategories()
	fmt.Println(cs)
}

func readCategories() []Category {
	doc, err := goquery.NewDocument("https://www.hackerrank.com/domains")
	if err != nil {
		log.Fatal(err)
	}

	var result []Category

	// read javascript object that contains the information on categories
	re := regexp.MustCompile(`HR\.PREFETCH_DATA\s*=\s*({.*});`)

	doc.Find("script").EachWithBreak(func(i int, s *goquery.Selection) bool {
		res := re.FindStringSubmatch(s.Text())
		if len(res) > 0 {
			jsonBlob := []byte(res[1])

			var dat struct {
				Contest struct {
					Categories []Category
				}
			}
			err := json.Unmarshal(jsonBlob, &dat)
			if err != nil {
				log.Fatal(err)
			}
			result = dat.Contest.Categories
			return false // stop `Each` iteration
		}
		return true
	})
	return result
}
