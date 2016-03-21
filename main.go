// TODO: use selenium

package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"regexp"

	"github.com/PuerkitoBio/goquery"
)

func check(e error) {
	if e != nil {
		log.Fatal(e)
	}
}

// Constants
const CHALLENGE_BASEURL = "http://www.hackerrank.com/challenges/"
const CATEGORY_BASEURL = "http://www.hackerrank.com/domains/"
const CATEGORIES_FILE = "categories.json"
const (
	EASY     = iota
	MEDIUM   = iota
	HARD     = iota
	ADVANCED = iota
	EXPERT   = iota
)

type Difficulty uint

type Challenge struct {
	Name, Slug string
	Difficulty Difficulty
	Category   Category
}

type Category struct {
	Name, Slug string
	Super      *Category  `json:"-"`
	Categories []Category `json:"children"`
}

func (c Category) String() string {
	p := "<nil>"
	if c.Super != nil {
		p = c.Super.Slug
	}
	return fmt.Sprintf("{%v %v %v %v}", c.Name, c.Slug, p, c.Categories)
}

func (c *Category) getFullSlug() string {
	res := ""
	for it := c; it != nil; it = it.Super {
		res = it.Slug + "/" + res
	}
	return res
}

func (c *Category) assignSuper(p *Category) {
	c.Super = p
	for i := range c.Categories {
		c.Categories[i].assignSuper(c)
	}
}

func (c *Category) parseChallenges() []Challenge {
	url := CATEGORY_BASEURL + c.getFullSlug()
	doc, err := goquery.NewDocument(url)
	check(err)

	// TODO: not working (HR is using javascript to load the content)
	// doc.Find(".content--list track_content").EachWithBreak(func(i int, s *goquery.Selection) bool {
	// 	// fmt.Println("s")
	// 	return false
	// })
	return nil
}

func main() {
	// check if categories were parsed in the past
	var cs []Category
	if _, err := os.Stat(CATEGORIES_FILE); err == nil {
		cs = readCategories(CATEGORIES_FILE)
	} else {
		cs = parseCategories()
		saveCategories(cs, CATEGORIES_FILE)
	}

	// try to parse on category
	cs[1].Categories[0].parseChallenges()

	// for i := range cs {
	// 	for j := range cs[i].Categories {
	// 		cs[i].Categories[j].parseChallenges()
	// 	}
	// }
}

func saveCategories(cs []Category, filename string) {
	// to JSON format
	b, err := json.MarshalIndent(cs, "", "\t")
	check(err)

	err = ioutil.WriteFile(filename, b, 0755)
	check(err)
}

func readCategories(filename string) []Category {
	data, err := ioutil.ReadFile(filename)
	check(err)

	var result []Category
	err = json.Unmarshal(data, &result)
	check(err)

	// assign super categories
	for i := range result {
		result[i].assignSuper(nil)
	}
	return result
}

func parseCategories() []Category {
	doc, err := goquery.NewDocument("https://www.hackerrank.com/domains")
	check(err)

	var result []Category

	// read javascript object that contains the information on categories
	re := regexp.MustCompile(`HR\.PREFETCH_DATA\s*=\s*({.*});`)

	doc.Find("script").EachWithBreak(func(i int, s *goquery.Selection) bool {
		res := re.FindStringSubmatch(s.Text())
		if len(res) > 0 {
			jsonBlob := []byte(res[1])

			// create variable to match the structure of the javascript object
			// that contains the necessary information
			var dat struct {
				Contest struct {
					Categories []Category
				}
			}
			err := json.Unmarshal(jsonBlob, &dat)
			check(err)

			// get result and assign parent categories to each
			// subcategory recursively
			result = dat.Contest.Categories
			for i := range result {
				result[i].assignSuper(nil)
			}
			return false // stop `EachWithBreak()` iteration
		}
		return true
	})
	return result
}
