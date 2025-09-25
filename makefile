up:
	docker-compose up -d

b:
	docker-compose build
d:
	docker-compose down

app:
	docker-compose up api-backend
appb:
	docker-compose up --build api-backend
f:
	docker-compose up mindmap-frontend

m:
	docker-compose up migrate
c:
	docker-compose up caddy

test:
	docker exec -it $(API) go test ./...
t:
	docker exec -it $(API) go test ./repo/...

mig:
	docker exec -it $(API) migrate -source file://repository/migrations -database postgres://postgres:password@postgres:5432/test?sslmode=disable up

td:
	docker exec -it -w /app/$(DIR) $(API) go test -cover

crud:
	docker exec -it  $(API) go run scripts/gen/main.go
repo:
	docker exec -it $(API) go run scripts/gen/makeRepository.go -model=$(MODEL) -table=$(shell echo $(MODEL) | tr 'A-Z' 'a-z')s
##	docker exec -it -w /app/auth mindmap-api go test -cover

# Значение по умолчанию
DIR ?= auth
MODEL ?= User
API ?= mindmap-api