FROM golang:1.11 AS builder
WORKDIR $GOPATH/src/gitlab.com/aido93/latex-server
COPY . .
RUN mkdir /uploads && go get -d -v ./...
RUN go install -v ./... && which latex-server

FROM blang/latex:ubuntu
LABEL maintainer="Igor Diakonov <aidos.tanatos@gmail.com>"
COPY --from=builder /go/bin/latex-server /latex-server
RUN apt-get update && \
    apt-get install -y --no-install-recommends texlive-lang-cyrillic texlive-fonts-extra && \
    rm -rf /var/lib/apt/lists/* && rm -rf /var/cache/apt/archives/
EXPOSE 8080
ENV CALLBACK_URL="" \
    DEBUG="false" \
    LOGLEVEL="info"
CMD ["/latex-server"]

