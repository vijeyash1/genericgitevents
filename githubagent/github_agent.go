package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"net/http"

	"github.com/go-git/go-git/v5"
	"github.com/go-playground/webhooks/v6/github"
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

		hook, _ := github.New(github.Options.Secret("helloworld"))
		payload, err := hook.Parse(r, github.PushEvent)
		//fmt.Printf("%T \n", payload)

		if err != nil {
			if err == github.ErrEventNotFound {
				// ok event wasn;t one of the ones asked to be parsed
				fmt.Println("this event was not present")
			}
		}
		switch payload.(type) {

		case github.PushPayload:
			release := payload.(github.PushPayload)
			var url string = release.Repository.HTMLURL
			var by, at, repo string = release.Commits[0].Author.Name, release.HeadCommit.Timestamp, release.Repository.Name
			publishGithubMetrics(url, by, at, repo, js)

			//fmt.Printf("commited by: %s \n", release.Commits[0].Author.Name)
			//fmt.Printf("commited at: %s \n", release.HeadCommit.Timestamp)
			//fmt.Printf("repository name:%s \n", release.Repository.Name)
			//	fmt.Printf("default branch:%s \n", release.Repository.DefaultBranch)
			//	fmt.Printf("master branch:%s \n", release.Repository.MasterBranch)
		}
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

func publishGithubMetrics(url string, by string, at string, repo string, js nats.JetStreamContext) (bool, error) {
	metrics := model.Githubevent{
		CommitedBy: by,
		CommitedAt: at,
		Repository: repo,
	}
	dir, err := ioutil.TempDir("", "commit-stat")
	checkErr(err)
	defer os.RemoveAll(dir) // clean up
	r, err := git.PlainClone(dir, false, &git.CloneOptions{
		URL: url,
	})
	checkErr(err)
	// ... retrieving the branch being pointed by HEAD
	ref, err := r.Head()
	checkErr(err)
	// ... retrieving the commit object
	commit, err := r.CommitObject(ref.Hash())
	checkErr(err)
	stats, _ := commit.Stats()

	for _, stat := range stats {
		fmt.Printf("add: %d lines\tdel: %d lines\tfile: %s\n", stat.Addition, stat.Deletion, stat.Name)
		metrics.Added = stat.Addition
		metrics.Deleted = stat.Deletion
		metrics.Filename = stat.Name
	}
	metricsJson, _ := json.Marshal(metrics)
	_, err = js.Publish(eventSubject, metricsJson)
	if err != nil {
		return true, err
	}
	fmt.Println(string(metricsJson))
	log.Printf("Metrics with repo name:%s has been published\n", repo)
	return false, nil
}
