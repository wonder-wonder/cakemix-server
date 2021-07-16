DBHOST=localhost
DBPORT=5432
DBUSER=postgres
DBPASS=postgres
DBNAME=cakemix
VERSION=$(shell git describe --tags)

rundev: main.go
	DBHOST="$(DBHOST)" DBPORT="$(DBPORT)" DBUSER="$(DBUSER)" DBPASS="$(DBPASS)" DBNAME="$(DBNAME)" go run -ldflags "-X main.version=$(VERSION)" -race main.go -c example/cakemix.conf.dev

test: main.go
	mkdir -p out/cover
	DBHOST="$(DBHOST)" DBPORT="$(DBPORT)" DBUSER="$(DBUSER)" DBPASS="$(DBPASS)" DBNAME="$(DBNAME)" go test -ldflags "-X main.version=$(VERSION)" -v ./handler -count=1 -coverprofile=out/cover/cover.out
	go tool cover -html=out/cover/cover.out -o out/cover/cover.html

startdb:
	docker run -dp $(DBPORT):5432 -v `pwd`/docker/postgres/init:/docker-entrypoint-initdb.d --name cakemixdbdev -e POSTGRES_PASSWORD=postgres postgres

stopdb:
	docker stop cakemixdbdev
	docker rm cakemixdbdev

runprod:
	docker-compose up --build -d

down:
	docker-compose down

build: main.go
	CGO_ENABLED=0 go build -o cakemixsv -ldflags "-X main.version=$(VERSION)" main.go

cleanall:
	rm -rf out
	rm -rf docker/server