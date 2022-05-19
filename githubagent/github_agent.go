package main

import (
	"fmt"

	"net/http"

	"github.com/go-playground/webhooks/v6/github"
)

const (
	path = "/webhooks"
)

func event(w http.ResponseWriter, r *http.Request) {
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
		// Do whatever you want from here...
		//fmt.Printf("%+v", release)
		//fmt.Println("data")

		// for _, val := range release.Commits {
		// 	for _, v := range val.Modified {
		// 		fmt.Printf("added: %s \n", v)
		// 	}
		// 	fmt.Printf("author name: %s \n", val.Author.Name)
		// }
		fmt.Printf("commited by: %s \n", release.Commits[0].Author.Name)
		fmt.Printf("commited at: %s \n", release.HeadCommit.Timestamp)
		fmt.Printf("repository name:%s \n", release.Repository.Name)

	}
}

func main() {

	http.HandleFunc(path, event)
	fmt.Println("github webhook server started at port 8000")
	http.ListenAndServe(":8000", nil)
}
