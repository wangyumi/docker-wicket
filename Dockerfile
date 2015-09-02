#FROM reg.docker.alibaba-inc.com/golang:1.4
#MAINTAINER tgic <farmer1992@gmail.com>

#RUN   mkdir -p /go/src/golang.org/x && cd /go/src/golang.org/x  && git clone https://github.com/golang/net
#COPY . /go/src/github.com/tg123/docker-wicket
#RUN go get github.com/tg123/docker-wicket

FROM reg.docker.alibaba-inc.com/ubuntu
MAINTAINER wangyumi <wangyumi@gamil.com>

EXPOSE 9999
COPY ./docker-wicket /docker-wicket

ENTRYPOINT ["/docker-wicket"]

CMD ["-h"]
