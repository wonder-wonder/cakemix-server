# Build server
FROM golang AS serverbuildenv
WORKDIR /go/src/github.com/wonder-wonder/cakemix-server
COPY . .
RUN make build

# Build front
FROM node:14-stretch AS frontbuildenv
RUN cd / && git clone https://github.com/wonder-wonder/cakemix-front
WORKDIR /cakemix-front
RUN mv config.example config && sed 's/wss/ws/g;s/https/http/g' -i config/.env.prod
RUN npm ci && npm run generate

# Construct publish image
FROM alpine
COPY example/cakemix.conf.prod /etc/cakemix/cakemix.conf
COPY share/mail /usr/share/cakemix/mail
COPY --from=serverbuildenv /go/src/github.com/wonder-wonder/cakemix-server/cakemixsv /usr/bin/cakemix
COPY --from=frontbuildenv /cakemix-front/dist /usr/share/cakemix/www

CMD ["/usr/bin/cakemix"]