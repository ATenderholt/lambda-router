# Start by building the application.
FROM golang:1.17-bullseye as build

WORKDIR /go/src/app
ADD . /go/src/app

RUN go get -d -v ./...

RUN go build -o /go/bin/app /go/src/app/cmd/functions

# Now copy it into our base image.
FROM gcr.io/distroless/base-debian11:debug
COPY --from=build /go/bin/app /

ENV NAME=rainbow-functions
EXPOSE 9050

ENTRYPOINT ["/app", "-local=false", "-data-path", "/data"]
