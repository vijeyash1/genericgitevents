package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"net/http"

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
			var by, at, repo string = release.Commits[0].Author.Name, release.HeadCommit.Timestamp, release.Repository.Name
			publishGithubMetrics(by, at, repo, js)

			fmt.Printf("commited by: %s \n", release.Commits[0].Author.Name)
			fmt.Printf("commited at: %s \n", release.HeadCommit.Timestamp)
			fmt.Printf("repository name:%s \n", release.Repository.Name)
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

func publishGithubMetrics(by string, at string, repo string, js nats.JetStreamContext) (bool, error) {
	metrics := model.Githubevent{
		CommitedBy: by,
		CommitedAt: at,
		Repository: repo,
	}
	metricsJson, _ := json.Marshal(metrics)
	_, err := js.Publish(eventSubject, metricsJson)
	if err != nil {
		return true, err
	}
	log.Printf("Metrics with repo name:%s has been published\n", repo)
	return false, nil
}
