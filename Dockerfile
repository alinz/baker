# build stage
FROM golang:alpine AS build-env
RUN apk --no-cache add build-base git bzr mercurial gcc
ADD . /src
RUN cd /src && GOGC=off GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-w -s" -o baker ./cmd/baker/main.go

# final stage
FROM alpine
WORKDIR /app
COPY --from=build-env /src/baker /app/baker

EXPOSE 80
EXPOSE 443

ENTRYPOINT ./baker