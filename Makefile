rundev: main.go
	DBHOST="localhost" DBPORT="5432" DBUSER="postgres" DBPASS="postgres" DBNAME="cakemix" APIADDR="localhost" PORT="8081" go run main.go

startdb:
	docker run -dp 5432:5432 -v `PWD`/docker/postgres:/docker-entrypoint-initdb.d --name cakemixdbdev -e POSTGRES_PASSWORD=postgres postgres

stopdb:
	docker stop geekersdbdev
	docker rm geekersdbdev

# runprod:
# 	docker-compose up

# dockerdown:
# 	docker-compose down

# build: main.go
# 	go build -o geekerssv main.go

key:
	yes|ssh-keygen -t rsa -f signkey -m PEM -N ""
	ssh-keygen -f signkey.pub -e -m pkcs8 > signkey.pub2
	mv signkey.pub2 signkey.pub
