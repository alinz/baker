<p align="center">
  <img height="150" src="https://github.com/alinz/baker/raw/master/logo.png"/>
</p>

# Baker

baker is a dynamic reverse proxy, which routes and load balacing incoming traffics to sets of containers. It is highly extensible and has set of rich middleweres and rules

### Features

- Highly extensible
- Support ACME TLS out of box (Let's encrypt)
- Dynamic configuration
- Support Round Robin load balancing
- Uses only go standrad libraies

### Usage

- as library

```
go get -u github.com/alinz/baker
```

- as docker image

```
docker pull alinz/baker:latest
```

- run baker as docker-compose

```yml
version: '3.5'

services:
  service1:
    image: alinz/baker:latest

    environment:
      # enables ACME system
      - BAKER_ACME=false
      # folder location which holds all certification
      - BAKER_ACME_PATH=/acme/cert

    ports:
      - '80:80'
      - '443:443'

    # make sure to use the right network
    networks:
      - baker

    volumes:
      # make sure it can access to main docker.sock
      - /var/run/docker.sock:/var/run/docker.sock
      - ./acme/cert:/acme/cert

networks:
  baker:
    name: baker_net
    driver: bridge
```

- run each container inside docker-compose

```yml
version: '3.5'

services:
  service1:
    image: my/service:latest

    labels:
      # this referes to baker's network
      # this is a key for baker to find which network
      # cotnainer is running
      - 'baker.network=baker_net'
      - 'baker.service.port=8000'
      # path to /config file in server
      - 'baker.service.ping=/config'

    networks:
      - baker

# make sure it references the baker's docker-compose network
networks:
  baker:
    external:
      name: baker_net
```
