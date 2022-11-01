include .env
export

dcu: ### Run docker-compose
	docker-compose up --build -d chat
.PHONY: dcu

dcd: ### Down docker-compose
	docker-compose down --remove-orphans
.PHONY: dcd