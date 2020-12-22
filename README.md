# Cakemix Server
Document management system for all creative team
Real-time edit with multiple users

## How To Run (for development)

``` sh
make key      # Generate public/private keys
make startdb  # Start database server on docker
make rundev   # Start server for development
# After testing
make stopdb   # Stop database server
```

## How To Run (for production)

``` sh
docker network create shared-network # Create network for public
make runprod                         # Build docker image and start
# When you want to stop server...
make down
```