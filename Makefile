.PHONY: bootstrap

bootstrap:
	go run . migrate up
	go run . seed roles
	go run . stripe sync
	go run . stripe role

.PHONY: reset
reset:
	go run . migrate reset
	go run . seed roles
	go run . stripe sync
	go run . stripe role

.PHONY: setupuser1
setupuser1:
	go run . seed user tkahng+01@gmail.com Password123! true
	go run . seed team tkahng+01@gmail.com team1

.PHONY: gotestsum
gotestsum:
	gotestsum -- -v ./... -race

.PHONY: gotest
gotest:
	go test -v ./... -race