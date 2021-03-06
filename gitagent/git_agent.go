package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"net/http"

	billy "github.com/go-git/go-billy/v5"
	memfs "github.com/go-git/go-billy/v5/memfs"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	htt "github.com/go-git/go-git/v5/plumbing/transport/http"
	memory "github.com/go-git/go-git/v5/storage/memory"
	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
	"github.com/vijeyash1/genericgitevents/model"
)

const (
	streamName     = "GITMETRICS"
	streamSubjects = "GITMETRICS.*"
	eventSubject   = "GITMETRICS.event"
	allSubject     = "GITMETRICS.all"
	path           = "/webhooks"
)

var (
	token    = os.Getenv("NATS_TOKEN")
	natsurl  = os.Getenv("NATS_ADDRESS")
	gituser  = os.Getenv("GIT_USER")
	gittoken = os.Getenv("GIT_TOKEN")
)

type Giturl struct {
	Repository struct {
		URL        string `json:"url"`
		GitHTTPURL string `json:"git_http_url"`
	}
}

type Branches []string

func (p Giturl) urlCheck() (g string) {
	url, url1 := p.Repository.URL, p.Repository.GitHTTPURL
	var u []string = []string{url, url1}
	refPrefix := "https"
	for _, ref := range u {
		if !strings.HasPrefix(ref, refPrefix) {
			continue
		}
		g = ref
	}
	return
}

var storer *memory.Storage
var fs billy.Filesystem

func main() {

	uuid := uuid.New()

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

		length := "https://github.com/"
		url := p.urlCheck()
		repo := url[len(length):]

		publishGithubMetrics(uuid, url, repo, gituser, gittoken, js)

	})
	fmt.Println("git webhook server started at port 8000")
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

func publishGithubMetrics(uuid uuid.UUID, url, repo, user, token string, js nats.JetStreamContext) (bool, error) {
	metrics := model.Githubevent{
		Repository: repo,
		Uuid:       uuid,
	}
	storer = memory.NewStorage()
	fs = memfs.New()
	// Authentication
	auth := &htt.BasicAuth{
		Username: user,
		Password: token,
	}
	r, err := git.Clone(storer, fs, &git.CloneOptions{
		URL:  url,
		Auth: auth,
	})
	checkErr(err)
	// dir, err := ioutil.TempDir("", "commit-stat")
	// checkErr(err)
	// defer os.RemoveAll(dir) // clean up
	// r, err := git.PlainClone(dir, false, &git.CloneOptions{
	// 	URL: url,
	// })
	// checkErr(err)
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

	var branches Branches
	for _, ref := range refList {

		refName := ref.Name().String()
		if !strings.HasPrefix(refName, refPrefix) {
			continue
		}
		branchName := refName[len(refPrefix):]

		branches = append(branches, branchName)

	}
	metrics.Availablebranches = totalbranches(&branches)
	// ... retrieving the branch being pointed by HEAD
	ref, err := r.Head()
	checkErr(err)
	// ... retrieving the commit object
	commit, err := r.CommitObject(ref.Hash())
	checkErr(err)

	metrics.CommitedBy = commit.Author.Name
	metrics.CommitedAt = commit.Author.When
	metrics.Commitmessage = commit.Message

	stats, _ := commit.Stats()

	metrics.Commitstat = getCommitStats(stats)

	metricsJson, _ := json.Marshal(metrics)
	_, err = js.Publish(eventSubject, metricsJson)
	if err != nil {
		return true, err
	}
	fmt.Println(string(metricsJson))
	log.Printf("Metrics with url:%s has been published\n", url)
	return false, nil
}

func getCommitStats(stat object.FileStats) string {
	var sb strings.Builder
	for _, comm := range stat {
		sb.WriteString(comm.Name)
		sb.WriteString(",")
		sb.WriteString("Add" + ":")
		sb.WriteString(fmt.Sprintf("%v", comm.Addition))
		sb.WriteString(",")
		sb.WriteString("Del" + ":")
		sb.WriteString(fmt.Sprintf("%v", comm.Deletion))
		sb.WriteString("  ")
	}
	return sb.String()
}

func totalbranches(b *Branches) string {
	var sb strings.Builder
	for _, bran := range *b {
		sb.WriteString(bran)
		sb.WriteString(",")
	}
	return sb.String()
}
