package news

import (
	"html"
	"net/http"
	"strings"
	"time"
	"unicode"

	"github.com/PuerkitoBio/goquery"
	twitterscraper "github.com/n0madic/twitter-scraper"
	log "github.com/sirupsen/logrus"
)

//TwitterTrends -
type TwitterTrends struct {
	Tweets   []tweet
	LastScan time.Time
}

type tweet struct {
	Subject string
	Date    time.Time
	News    []string
}

//GetTrends -
func (twitterTrends *TwitterTrends) GetTrends() {
	trends, err := twitterscraper.GetTrends()
	if err != nil {
		log.Warn("Twitter API not reachable", err)
		return
	}
	for _, a := range trends {
		err = twitterTrends.saveTweet(a)
		if err != nil {
			log.Error(err)
		}
	}
	twitterTrends.LastScan = time.Now()
}

func (twitterTrends *TwitterTrends) saveTweet(tweetSubject string) error {
	for i, a := range twitterTrends.Tweets {
		if a.Subject == tweetSubject {
			twitterTrends.Tweets[i].Date = time.Now()
			log.Debug(" %s already found and renewed", a.Subject)
			return nil
		}
	}

	response, err := googleRequest(tweetSubject)
	if err != nil {
		return err
	}
	results, err := googleResultParser(response)
	if err != nil {
		return err
	}

	newTweet := &tweet{
		Subject: tweetSubject,
		Date:    time.Now(),
		News:    results,
	}
	twitterTrends.Tweets = append(twitterTrends.Tweets, *newTweet)
	log.Info("Added new Trend: ", newTweet.Subject)
	return nil
}

//CleanOldTrends -
func (twitterTrends *TwitterTrends) CleanOldTrends() {
	log.Debug("Start Cleanup")
	currentDate := time.Now()
	for _, tweet := range twitterTrends.Tweets {
		if tweet.Date.Add(10 * time.Minute).Before(currentDate) {
			log.Info("Cleanup %s \n", tweet.Subject)
		}
	}
	log.Debug("Finished Cleanup")
}

func googleRequest(queryString string) (*http.Response, error) {
	queryString = strings.ReplaceAll(queryString, " ", "%20")
	queryString = strings.ReplaceAll(queryString, "#", "")

	searchURL := "https://www.google.com/search?q=" + queryString + "&hl=en&tbm=nws"
	baseClient := &http.Client{}

	req, _ := http.NewRequest("GET", searchURL, nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/61.0.3163.100 Safari/537.36")
	res, err := baseClient.Do(req)
	log.Debug("query: ", searchURL)

	if err != nil {
		return nil, err
	}
	return res, nil
}

func googleResultParser(response *http.Response) ([]string, error) {
	doc, err := goquery.NewDocumentFromResponse(response)
	if err != nil {
		return nil, err
	}
	var results []string
	cnt := 0
	doc.Find("div").Each(func(i int, s *goquery.Selection) {
		band, ok := s.Attr("role")
		if ok {
			if band == "heading" {
				if cnt >= 3 {
					return
				}
				cnt++
				title, _ := s.Html()
				title = strings.Map(func(r rune) rune {
					if unicode.IsGraphic(r) {
						return r
					}
					return -1
				}, title)
				title = strings.Trim(title, "...")
				title = html.UnescapeString(title)
				results = append(results, title)
			}
		}
	})
	log.Debug("found: ", cnt)
	return results, err
}
