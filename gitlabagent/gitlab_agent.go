package main

import (
	"fmt"

	"net/http"

	"github.com/go-playground/webhooks/v6/gitlab"
)

const (
	path = "/webhooks"
)

func main() {
	hook, _ := gitlab.New(gitlab.Options.Secret("helloworld"))

	http.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		payload, err := hook.Parse(r, gitlab.PushEvents)
		if err != nil {
			if err == gitlab.ErrEventNotFound {
				// ok event wasn;t one of the ones asked to be parsed
				fmt.Println("this event was not present")
			}
		}
		switch payload.(type) {

		case gitlab.PushEventPayload:
			release := payload.(gitlab.PushEventPayload)
			// Do whatever you want from here...
			//fmt.Printf("%+v", release)
			fmt.Printf("%s", release.UserName)
		}
	})
	fmt.Println("gitlab webhook server started at port 8000")
	http.ListenAndServe(":8000", nil)
}
