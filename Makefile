DBHOST=localhost
DBPORT=5432
DBUSER=postgres
DBPASS=postgres
DBNAME=cakemix
VERSION=$(shell git describe --tags)

rundev: main.go out/keys/signkey
	DBHOST="$(DBHOST)" DBPORT="$(DBPORT)" DBUSER="$(DBUSER)" DBPASS="$(DBPASS)" DBNAME="$(DBNAME)" go run -ldflags "-X main.version=$(VERSION)" -race main.go -c example/cakemix.conf.dev

test: main.go signkey
	test -d out || mkdir out
	DBHOST="$(DBHOST)" DBPORT="$(DBPORT)" DBUSER="$(DBUSER)" DBPASS="$(DBPASS)" DBNAME="$(DBNAME)" go test -ldflags "-X main.version=$(VERSION)" -v ./handler -count=1 -coverprofile=out/cover.out
	go tool cover -html=out/cover.out -o out/cover.html

startdb:
	docker run -dp $(DBPORT):5432 -v `pwd`/docker/postgres/init:/docker-entrypoint-initdb.d --name cakemixdbdev -e POSTGRES_PASSWORD=postgres postgres

stopdb:
	docker stop cakemixdbdev
	docker rm cakemixdbdev

runprod: docker/server/keys/signkey
	docker-compose up --build -d

down:
	docker-compose down

build: main.go
	go build -o cakemixsv -ldflags "-X main.version=$(VERSION)" main.go

out/keys/signkey:
	mkdir -p out/keys
	ssh-keygen -t rsa -f out/keys/signkey -m PEM -N ""
	ssh-keygen -f out/keys/signkey.pub -e -m pkcs8 > out/keys/signkey.pub2
	mv out/keys/signkey.pub2 out/keys/signkey.pub

docker/server/keys/signkey:
	mkdir -p docker/server/keys
	cd docker/server/keys &&\
	make -f ../../../Makefile signkey

cleanall:
	rm -rf out
	rm -f signkey
	rm -f signkey.pub
	rm -rf cmdat
	rm -rf docker/server