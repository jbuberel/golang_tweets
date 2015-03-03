package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/jbuberel/anaconda"
)

var twitterConsumerKey string = ""
var twitterConsumerSecret string = ""
var twitterAccessToken string = ""
var twitterSecretToken string = ""
var since *string
var until *string

func init() {
	log.SetFlags(log.Ltime | log.Lshortfile)

	log.Printf("Looking through env vars\n")
	for _, e := range os.Environ() {
		parts := strings.Split(e, "=")
		if len(parts) == 2 {
			if parts[0] == "twitter_consumer_key" {
				twitterConsumerKey = string(parts[1])
				log.Printf("twitter_consumer_key set from environ to: %v\n", twitterConsumerKey)
			} else if parts[0] == "twitter_consumer_secret" {
				twitterConsumerSecret = string(parts[1])
				log.Printf("twitter_consumer_secret set from environ to: %v\n", twitterConsumerSecret)
			} else if parts[0] == "twitter_access_token" {
				twitterAccessToken = string(parts[1])
				log.Printf("twitter_access_token set from environ to: %v\n", twitterAccessToken)
			} else if parts[0] == "twitter_secret_token" {
				twitterSecretToken = string(parts[1])
				log.Printf("twitter_secret_token set from environ to: %v\n", twitterSecretToken)
			}
		}
	}

	since = flag.String("since", time.Now().Add(-47*time.Hour).Format("2006-01-02"), "The start date of the search window, in YYYY-MM-DD format.")
	until = flag.String("until", time.Now().Add(-23*time.Hour).Format("2006-01-02"), "The end date of the search window, in YYYY-MM-DD format.")
	flag.Parse()
}

func extract(api *anaconda.TwitterApi, term string) map[string]anaconda.Tweet {
	log.Printf("Beginning tweet extraction for date range %v to %v.\n", *since, *until)
	v := url.Values{}
	v.Add("result_type", "recent")

	tweets := make(map[string]anaconda.Tweet)

	q := fmt.Sprintf("%v since:%v until:%v", term, *since, *until)
	log.Printf("q: %v\n", q)
	for searchResult, _ := api.GetSearch(q, v); len(searchResult.Statuses) > 0; searchResult, _ = searchResult.GetNext(api) {
		for _, tweet := range searchResult.Statuses {
			tweets[tweet.IdStr] = tweet
		}
		log.Printf("Total tweets %v\n", len(tweets))
		time.Sleep(5 * time.Second)
	}

	log.Printf("Completing tweet extraction, found %v tags and %v mentions.", len(tweets), 0)
	return tweets

}

func main() {
	log.Printf("Connecting to twitter\n")
	anaconda.SetConsumerKey(twitterConsumerKey)
	anaconda.SetConsumerSecret(twitterConsumerSecret)
	api := anaconda.NewTwitterApi(twitterAccessToken, twitterSecretToken)
	tweets := extract(api, "#golang")

	tweetJSON, err := json.Marshal(tweets)
	if err != nil {
		fmt.Printf("Unable to marshal tweets to JSON %v", err)
		return
	}
	err = ioutil.WriteFile(fmt.Sprintf("tags-golang-%v", *until), tweetJSON, 0644)
	if err != nil {
		log.Printf("Unable to write file %v", err)
	}

	tweets = extract(api, "@golang")
	tweetJSON, err = json.Marshal(tweets)
	if err != nil {
		fmt.Printf("Unable to marshal tweets to JSON %v", err)
		return
	}
	err = ioutil.WriteFile(fmt.Sprintf("mentions-golang-%v", *until), tweetJSON, 0644)
	if err != nil {
		log.Printf("Unable to write file %v", err)
	}

}
