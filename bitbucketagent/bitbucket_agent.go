package main

import (
	"fmt"

	"net/http"

	"github.com/go-playground/webhooks/v6/bitbucket"
)

const (
	path = "/webhooks"
)

func main() {
	hook, _ := bitbucket.New()

	http.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		payload, err := hook.Parse(r, bitbucket.RepoPushEvent)
		if err != nil {
			if err == bitbucket.ErrEventNotFound {
				// ok event wasn;t one of the ones asked to be parsed
				fmt.Println("this event was not present")
			}
		}
		switch payload.(type) {

		case bitbucket.RepoPushPayload:
			release := payload.(bitbucket.RepoPushPayload)
			// Do whatever you want from here...
			//fmt.Printf("%+v", release)
			fmt.Printf("%s", release.Actor.DisplayName)
		}
	})
	fmt.Println("bitbucket webhook server started at port 8000")
	http.ListenAndServe(":8000", nil)
}
