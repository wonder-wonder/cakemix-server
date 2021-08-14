# Build server
FROM golang AS serverbuildenv
WORKDIR /go/src/github.com/wonder-wonder/cakemix-server
COPY . .
RUN make build

# Construct publish image
FROM alpine
COPY example/cakemix.conf.prod /etc/cakemix/cakemix.conf
COPY share/mail /usr/share/cakemix/mail
COPY --from=serverbuildenv /go/src/github.com/wonder-wonder/cakemix-server/cakemixsv /usr/bin/cakemix

CMD ["/usr/bin/cakemix"]