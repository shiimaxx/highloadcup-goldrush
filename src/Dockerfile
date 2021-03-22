FROM golang:latest as build-env

ENV GOOS=linux
ENV GOARCH=amd64

WORKDIR /go/src/app
ADD . /go/src/app

RUN go build -o /go/bin/app

FROM gcr.io/distroless/base
COPY --from=build-env /go/bin/app /
CMD ["/app"]

