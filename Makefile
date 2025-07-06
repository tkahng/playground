.PHONY: bootstrap
bootstrap:
	go run . migrate up
	go run . seed roles
	go run . superuser create admin@k2dv.io Password123! 
	go run . stripe sync
	go run . stripe role

.PHONY: setupuser1
setupuser1:
	go run . seed user tkahng+01@gmail.com Password123! true
	go run . seed team tkahng+01@gmail.com team1

.PHONY: migrate-up
migrate-up:
	go run . migrate up

.PHONY: migrate-down
migrate-reset:
	go run . migrate reset
