FROM golang

WORKDIR /go/src/github.com/wonder-wonder/cakemix-server

COPY . .
COPY example/cakemix.conf.prod /etc/cakemix/cakemix.conf
COPY share/mail /usr/share/cakemix/mail

RUN make build

CMD ["./cakemixsv"]