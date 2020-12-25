FROM golang

WORKDIR /root/src/github.com/wonder-wonder/cakemix-server

COPY . .

ENV GOPATH /go:/root

RUN make build

CMD ["./cakemixsv"]