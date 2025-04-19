.PHONY: bootstrap
bootstrap:
	go run . seed roles
	go run . superuser create tkahng+01@gmail.com Password123! 
	go run . stripe sync
	go run . stripe role

.PHONY: migrate-up
migrate-up:
	go run . migrate up

.PHONY: migrate-down
migrate-reset:
	go run . migrate reset
