FROM golang

WORKDIR /go/src/github.com/wonder-wonder/cakemix-server

COPY . .

# ENV GOPATH /go:/root

RUN make build

CMD ["./cakemixsv","-c","cakemix.conf.prod"]