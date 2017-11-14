package main

import (
	"flag"
	"github.com/microcosm-cc/bluemonday"
	"github.com/minya/gopushover"
	"github.com/mmcdole/gofeed"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"time"
)

var url string
var pattern *regexp.Regexp
var poSettings *gopushover.PushoverSettings

func init() {
	flag.StringVar(&url, "url", "", "Url of the feed")
	var strPattern string
	flag.StringVar(&strPattern, "pattern", "", "Pattern to search")
	var logPath string
	flag.StringVar(&logPath, "logpath", "rsswatch.log", "Path to write logs")
	flag.Parse()
	if url == "" {
		flag.Usage()
		os.Exit(-1)
	}
	if strPattern == "" {
		flag.Usage()
		os.Exit(-1)
	}

	p, err := regexp.Compile(strPattern)
	if err != nil {
		os.Stderr.WriteString("Pattern should be a valid regex\n")
		os.Exit(-1)
	}
	pattern = p

	setUpLogger(logPath)

	settings, err := gopushover.ReadSettings("pushover.json")
	if err != nil {
		os.Stderr.WriteString("Can't get pushover settings\n")
		os.Exit(-1)
	}
	poSettings = &settings
}

func main() {
	log.Println("Start")
	parser := gofeed.NewParser()
	feed, err := parser.ParseURL(url)
	if err != nil {
		log.Printf("Unable to get feed from %v\n", url)
		os.Exit(-1)
	}

	date := readState()
	if err != nil {
		log.Printf("Error: %v\n", err)
		os.Exit(-1)
	}

	for _, item := range feed.Items {
		if pattern.MatchString(item.Title) && item.PublishedParsed.After(date) {
			log.Printf("Found an item on %v\n", date)
			saveState(item.PublishedParsed)
			notify(item)
			return
		}
	}

	log.Printf("Nothing was found")
}

func notify(item *gofeed.Item) {
	p := bluemonday.StripTagsPolicy()
	description := p.Sanitize(item.Description)
	gopushover.SendMessage(poSettings.Token, poSettings.User, item.Title, description)
}

func readState() time.Time {
	file, err := os.Open("state")
	if err != nil {
		return time.Now().AddDate(0, 0, -7)
	}
	data, err := ioutil.ReadAll(file)
	timeRead, err := time.Parse("2006-01-02 15:04:05", string(data))
	if err != nil {
		return time.Now().AddDate(0, 0, -7)
	}
	return timeRead
}

func saveState(itemTime *time.Time) {
	file, _ := os.OpenFile("state", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	file.WriteString(itemTime.Format("2006-01-02 15:04:05"))
}

func setUpLogger(logPath string) {
	logFile, err := os.OpenFile(logPath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	log.SetOutput(logFile)
}
