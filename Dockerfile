# Used for local testing only, since OSX support isn't super great right now
# OSXFUSE 3.2.0 starts but doesnt ReadDirAll for some reason.
# OSXFUSE 3.4.1 is busted altogether, mount errors due to args provided.

FROM golang:1.6.3-wheezy

RUN apt-get update && DEBIAN_FRONTEND=noninteractive apt-get install -y fuse

RUN mkdir -p /go/src/app
WORKDIR /go/src/app

RUN mkdir mountpoint

COPY ./*.go /go/src/app/

RUN go-wrapper download
RUN go-wrapper install

COPY ./test_cmd.sh /

CMD ["/test_cmd.sh"]
