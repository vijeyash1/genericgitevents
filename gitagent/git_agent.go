package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"net/http"

	"github.com/go-git/go-git/v5"
	"github.com/nats-io/nats.go"
	"github.com/vijeyash1/genericgitevents/model"
)

const (
	streamName     = "METRICS"
	streamSubjects = "METRICS.*"
	eventSubject   = "METRICS.event"
	allSubject     = "METRICS.all"
	path           = "/webhooks"
)

var token string = os.Getenv("NATS_TOKEN")
var natsurl string = os.Getenv("NATS_ADDRESS")

type Giturl struct {
	Repository struct {
		URL        string `json:"url"`
		GitHTTPURL string `json:"git_http_url"`
	}
}

func main() {

	// Connect to NATS
	nc, err := nats.Connect(natsurl, nats.Name("Github metrics"), nats.Token(token))
	checkErr(err)
	// Creates JetStreamContext
	js, err := nc.JetStream()
	checkErr(err)
	// Creates stream
	err = createStream(js)
	checkErr(err)
	http.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		var p Giturl

		err := json.NewDecoder(r.Body).Decode(&p)
		checkErr(err)

		url := p.Repository.URL

		length := "https://github.com/"
		repo := url[len(length):]

		publishGithubMetrics(url, repo, js)

	})
	fmt.Println("github webhook server started at port 8000")
	http.ListenAndServe(":8000", nil)
}
func checkErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

// createStream creates a stream by using JetStreamContext
func createStream(js nats.JetStreamContext) error {
	// Check if the METRICS stream already exists; if not, create it.
	stream, err := js.StreamInfo(streamName)
	log.Printf("Retrieved stream %s", fmt.Sprintf("%v", stream))
	if err != nil {
		log.Printf("Error getting stream %s", err)
	}
	if stream == nil {
		log.Printf("creating stream %q and subjects %q", streamName, streamSubjects)
		_, err = js.AddStream(&nats.StreamConfig{
			Name:     streamName,
			Subjects: []string{streamSubjects},
		})
		checkErr(err)
	}
	return nil
}

func publishGithubMetrics(url, repo string, js nats.JetStreamContext) (bool, error) {
	metrics := model.Githubevent{
		Repository: repo,
	}

	dir, err := ioutil.TempDir("", "commit-stat")
	checkErr(err)
	defer os.RemoveAll(dir) // clean up
	r, err := git.PlainClone(dir, false, &git.CloneOptions{
		URL: url,
	})
	checkErr(err)
	//"origin" is a shorthand name
	// for the remote repository that a project was originally cloned from
	remote, err := r.Remote("origin")
	if err != nil {
		panic(err)
	}
	refList, err := remote.List(&git.ListOptions{})
	if err != nil {
		panic(err)
	}
	refPrefix := "refs/heads/"
	for _, ref := range refList {
		refName := ref.Name().String()
		if !strings.HasPrefix(refName, refPrefix) {
			continue
		}
		branchName := refName[len(refPrefix):]
		metrics.Availablebranches = append(metrics.Availablebranches, branchName)

	}
	// ... retrieving the branch being pointed by HEAD
	ref, err := r.Head()
	checkErr(err)
	// ... retrieving the commit object
	commit, err := r.CommitObject(ref.Hash())
	checkErr(err)

	metrics.CommitedBy = commit.Author.Name
	metrics.CommitedAt = commit.Author.When

	stats, _ := commit.Stats()

	for _, stat := range stats {
		//fmt.Printf("add: %d lines\tdel: %d lines\tfile: %s\n", stat.Addition, stat.Deletion, stat.Name)
		metrics.Added = append(metrics.Added, stat.Addition)
		metrics.Deleted = append(metrics.Deleted, stat.Deletion)
		metrics.Filename = append(metrics.Filename, stat.Name)
	}
	metricsJson, _ := json.Marshal(metrics)
	_, err = js.Publish(eventSubject, metricsJson)
	if err != nil {
		return true, err
	}
	fmt.Println(string(metricsJson))
	log.Printf("Metrics with url:%s has been published\n", url)
	return false, nil
}
