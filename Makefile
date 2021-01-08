DBHOST="localhost"
DBPORT="5432"
DBUSER="postgres"
DBPASS="postgres"
DBNAME="cakemix"
APIADDR="localhost"
PORT="8081"

rundev: main.go
	test -f signkey || make key
	test ! -f sendgrid.env || . sendgrid.env
	DBHOST=$(DBHOST) DBPORT=$(DBPORT) DBUSER=$(DBUSER) DBPASS=$(DBPASS) DBNAME=$(DBNAME) APIADDR=$(APIADDR) PORT=$(PORT) go run -race main.go

test: main.go
	test -f signkey || make key
	DBHOST=$(DBHOST) DBPORT=$(DBPORT) DBUSER=$(DBUSER) DBPASS=$(DBPASS) DBNAME=$(DBNAME) APIADDR=$(APIADDR) PORT=$(PORT) go test -v ./handler --run=TestAuthHandler

startdb:
	docker run -dp 5432:5432 -v `pwd`/docker/postgres/init:/docker-entrypoint-initdb.d --name cakemixdbdev -e POSTGRES_PASSWORD=postgres postgres

stopdb:
	docker stop cakemixdbdev
	docker rm cakemixdbdev

runprod: keyprod
	docker-compose up --build -d

down:
	docker-compose down

build: main.go
	go build -o cakemixsv main.go

key:
	yes|ssh-keygen -t rsa -f signkey -m PEM -N ""
	ssh-keygen -f signkey.pub -e -m pkcs8 > signkey.pub2
	mv signkey.pub2 signkey.pub

keyprod:
	@test ! -f docker/server/keys/signkey &&\
	echo Generating signing keys &&\
	mkdir -p docker/server/keys &&\
	cd docker/server/keys &&\
	make -f ../../../Makefile key ||\
	echo Skipping key generation