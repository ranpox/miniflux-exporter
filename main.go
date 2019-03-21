package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"strconv"
	"time"

    miniflux "miniflux.app/client"

	humanize "github.com/dustin/go-humanize"
	"github.com/gorilla/feeds"
	"github.com/sirupsen/logrus"
)

var (
	targetOPMLFile     string
	targetBookmarkFile string
	username           string
	password           string
	hostname           string
	silent             bool
)

func main() {
	flag.StringVar(&targetOPMLFile, "output-opml", "", "output filename, f.e. /tmp/opml.xml")
	flag.StringVar(&targetBookmarkFile, "output-bookmarks", "", "output filename, f.e. /tmp/bookmarks.txt")
	flag.StringVar(&username, "user", "", "miniflux username")
	flag.StringVar(&password, "pass", "", "miniflux password")
	flag.StringVar(&hostname, "host", "http://localhost:8080", "miniflux hostname, f.e. http://localhost:8080")
	flag.BoolVar(&silent, "s", false, "if flag -s is provided, the happy-flow won't display any output")
	flag.Parse()

	// get miniflux client
	c := miniflux.New(hostname, username, password)

	// start export to opml
	if len(targetOPMLFile) > 0 {
		exportOPML(c)
	} else {
		message("skipping opml export (see -help for more info)")
	}

	// start export bookmarks/starred entries
	if len(targetBookmarkFile) > 0 {
		exportStarredEntries(c)
	} else {
		message("skipping export of bookmarks/starred entries (see -help for more info)")
	}

}

func exportOPML(c *miniflux.Client) {
	opmlFile, err := c.Export()
	if err != nil {
		logrus.Error(err)
	}

	err = ioutil.WriteFile(targetOPMLFile, opmlFile, 0644)
	if err != nil {
		logrus.Error(err)
		return
	}

	message(fmt.Sprintf("export OPML done, %s written to file %s", humanize.Bytes(uint64(len(opmlFile))), targetOPMLFile))
	return
}

func exportStarredEntries(c *miniflux.Client) {
	var (
		a      []byte
		number int
	)

	now := time.Now()
	feed := &feeds.Feed{
		Title:       "Miniflux starred entries",
		Description: "RSS feed from all starred entries in Miniflux",
		Link:        &feeds.Link{Href: hostname},
		Created:     now,
	}

	entries, err := c.Entries(&miniflux.Filter{})
	if err != nil {
		logrus.Error(err)
		return
	}

	for _, entry := range entries.Entries {
		if entry.Starred {

			newItem := feeds.Item{
				Title:       entry.Title,
				Link:        &feeds.Link{Href: entry.URL},
				Author:      &feeds.Author{Name: entry.Author},
				Description: entry.Content,
				Id:          strconv.Itoa(int(entry.ID)),
			}

			feed.Items = append(feed.Items, &newItem)
			number++
		}
	}

	rss, err := feed.ToRss()
	if err != nil {
		logrus.Errorf("error exporting starred items to RSS feed: %s", err)
		return
	}

	err = ioutil.WriteFile(targetBookmarkFile, []byte(rss), 0644)
	if err != nil {
		logrus.Error(err)
		return
	}

	message(fmt.Sprintf("export %d bookmarks done, %s written to file %s", number, humanize.Bytes(uint64(len(a))), targetBookmarkFile))
	return
}

func message(m string) {
	if !silent {
		logrus.Infof(m)
	}
}
