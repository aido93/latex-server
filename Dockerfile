FROM golang:1.11 AS builder
WORKDIR $GOPATH/src/gitlab.com/aido93/latex-server
COPY . .
RUN mkdir /uploads && go get -d -v ./...
RUN go install -v ./... && which latex-server

FROM blang/latex:ubuntu
LABEL maintainer="Igor Diakonov <aidos.tanatos@gmail.com>"
ENV LATEX_DIR=/usr/share/texlive/texmf-dist/tex/latex
COPY --from=builder /go/bin/latex-server /latex-server
RUN apt-get update -q && \
    apt-get install -y --no-install-recommends wget texlive-lang-cyrillic texlive-fonts-extra && \
    rm -rf /var/lib/apt/lists/* && rm -rf /var/cache/apt/archives/ && \
    mkdir -p $LATEX_DIR/pgf-pie/ && \
    wget http://mirrors.ctan.org/graphics/pgf/contrib/pgf-pie/pgf-pie.sty -O $LATEX_DIR/pgf-pie/pgf-pie.sty && \
    mktexlsr
EXPOSE 8080
ENV CALLBACK_URL="" \
    DEBUG="false" \
    LOGLEVEL="info"
CMD ["/latex-server"]

