preRequesties before running:
run a nats jetstream as docker

docker run -d -p 4222:4222 nats:latest -js

after running the docker...................

export these values:

export NATS_TOKEN="your nats token"(sample:"UfmrJOYwYCCsgQvxvcfJ3BdI6c8WBbnD")

export NATS_ADDRESS="nats://localhost:4222"

export "GIT_USER"="your git username"

export "GIT_TOKEN"="your git token"

After exporting and running nats js.............

gitevent can be run by :

go run gitevent.go

for testing purpose you can use hookdeck for creating webhook url.
you can also any similar ones.
