![Test Cakemix](https://github.com/wonder-wonder/cakemix-server/workflows/Test%20Cakemix/badge.svg?branch=main)

# Cakemix Server
Document management system for all creative team
Real-time edit with multiple users

# Installation
## Fast way (using docker)
Please prepare front data directory in project root as `dist` in advance.

``` sh
docker network create shared-network # Create network for public
make runprod                         # Build docker image and start
# When you want to stop server...
make down
```
Now you can access `localhost:8081` to use cakemix!

# For developer
## How To run for development
``` sh
make startdb  # Start database server on docker
make rundev   # Start server for development
# After testing
make stopdb   # Stop database server
```

# Envrionment variables
- Database
  - `DBHOST` is hostname for postgres database (default: `cakemixpg`)
  - `DBPORT` is port number for postgres database (default: `5432`)
  - `DBUSER` is user for postgres database (default: `postgres`)
  - `DBPASS` is password for postgres database (default: `postgres`)
  - `DBNAME` is database name for postgres database (default: `cakemix`)

- API
  - `APIADDR` is API server socket address (default: )
  - `PORT` is API server port (default: `8081`)

- Files and directories
  - `FRONTDIR` is directory for front files. If empty, the server only serve API. (default: )
  - `DATADIR` is directory for cakemix server variable data like uploaded pictures. (default: `./cmdat`)
  - `SIGNPRVKEY` is private key for signing token. (default: `./signkey`)
  - `SIGNPUBKEY` is public key for verification of token. (default: `./signkey.pub`)

- Mail
	- `SENDGRID_API_KEY` is SendGrid API Key. If `DEBUG` is specified, the content is showed in log. If empty, mail function is disabled. (default: )