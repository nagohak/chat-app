include .env
export

dcu: ### Run docker-compose
	docker-compose --env-file .env up --build -d chat
.PHONY: dcu

dcd: ### Down docker-compose
	docker-compose down --remove-orphans
.PHONY: dcd