FROM golang

WORKDIR /go/src/github.com/wonder-wonder/cakemix-server

COPY . .
COPY example/cakemix.conf.prod /etc/cakemix/cakemix.conf

RUN make build

CMD ["./cakemixsv"]