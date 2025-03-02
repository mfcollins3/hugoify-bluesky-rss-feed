// Copyright 2025 Michael F. Collins, III
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to
// deal in the Software without restriction, including without limitation the
// rights to use, copy, modify, merge, publish, distribute, sublicense, and/or
// sell copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING
// FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS
// IN THE SOFTWARE.

// Package main implements a transformation engine that will read an RSS feed
// from Blue Sky and will reformat the feed into a form that Hugo can use.
//
// The current problem with the Blue Sky RSS feed format is that the pubDate
// field is not formatted in a way that Hugo can parse the date and time from
// the pubDate field. This GitHub Action program will parse and rewrite the
// pubDate field into a format that Hugo can use.
package main

import (
	"encoding/xml"
	"log"
	"net/http"
	"os"
	"time"
)

type rss struct {
	XMLName xml.Name `xml:"rss"`
	Version string   `xml:"version,attr"`
	Channel channel  `xml:"channel"`
}

type channel struct {
	Description string `xml:"description"`
	Link        string `xml:"link"`
	Title       string `xml:"title"`
	Items       []item `xml:"item"`
}

type item struct {
	Link        string `xml:"link"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
	Guid        guid   `xml:"guid"`
}

type guid struct {
	IsPermaLink string `xml:"isPermaLink,attr"`
	Value       string `xml:",chardata"`
}

func main() {
	url, ok := os.LookupEnv("INPUT_URL")
	if !ok {
		log.Fatal("The url input is required.")
	}

	path, ok := os.LookupEnv("INPUT_PATH")
	if !ok {
		log.Fatal("The path input is required.")
	}

	resp, err := http.Get(url)
	if err != nil {
		log.Fatalf("Failed to download the RSS feed: %v", err)
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		log.Fatalf(
			"Failed to download RSS feed. Status code: %d",
			resp.StatusCode,
		)
	}

	var rss rss
	decoder := xml.NewDecoder(resp.Body)
	if err = decoder.Decode(&rss); err != nil {
		log.Fatalf("Failed to parse the RSS feed: %v", err)
	}

	for i := range rss.Channel.Items {
		pubDate, err := time.Parse(
			"02 Jan 2006 15:04 -0700",
			rss.Channel.Items[i].PubDate,
		)
		if err != nil {
			log.Fatalf("Failed to parse the pubDate field: %v", err)
		}

		rss.Channel.Items[i].PubDate = pubDate.Format(
			"2006-01-02T15:04:05-07:00",
		)
	}

	file, err := os.Create(path)
	if err != nil {
		log.Fatalf("Failed to create the file: %v", err)
	}

	defer func() {
		_ = file.Close()
	}()

	encoder := xml.NewEncoder(file)
	encoder.Indent("", "  ")
	if err = encoder.Encode(rss); err != nil {
		log.Fatalf("Failed to write the RSS feed: %v", err)
	}
}
