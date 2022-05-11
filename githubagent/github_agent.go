package main

import (
	"fmt"

	"net/http"

	"github.com/go-playground/webhooks/v6/github"
)

const (
	path = "/webhooks"
)

func main() {
	hook, _ := github.New(github.Options.Secret("helloworld"))

	http.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		payload, err := hook.Parse(r, github.PushEvent)

		if err != nil {
			if err == github.ErrEventNotFound {
				// ok event wasn;t one of the ones asked to be parsed
				fmt.Println("this event was not present")
			}
		}
		switch payload.(type) {

		case github.ReleasePayload:
			release := payload.(github.PushPayload)
			// Do whatever you want from here...
			//fmt.Printf("%+v", release)
			fmt.Printf("%s", release.Pusher.Name)
		}
	})
	fmt.Println("github webhook server started at port 8000")
	http.ListenAndServe(":8000", nil)
}
