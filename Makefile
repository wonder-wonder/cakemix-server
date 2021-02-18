DBHOST=localhost
DBPORT=5432
DBUSER=postgres
DBPASS=postgres
DBNAME=cakemix

rundev: main.go signkey sendgrid.env
	$(shell cat sendgrid.env) DBHOST="$(DBHOST)" DBPORT="$(DBPORT)" DBUSER="$(DBUSER)" DBPASS="$(DBPASS)" DBNAME="$(DBNAME)" go run -race main.go -c cakemix.conf

test: main.go signkey
	test -d out || mkdir out
	DBHOST="$(DBHOST)" DBPORT="$(DBPORT)" DBUSER="$(DBUSER)" DBPASS="$(DBPASS)" DBNAME="$(DBNAME)" go test -v ./handler -count=1 -coverprofile=out/cover.out
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
	go build -o cakemixsv main.go

signkey:
	ssh-keygen -t rsa -f signkey -m PEM -N ""
	ssh-keygen -f signkey.pub -e -m pkcs8 > signkey.pub2
	mv signkey.pub2 signkey.pub

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