FROM reg.docker.alibaba-inc.com/golang:1.4.2
MAINTAINER wangyumi <wangyumi@gamil.com>

RUN cd /go  && mkdir -p src/golang.org/x  && cd src/golang.org/x \
	&& git clone https://github.com/golang/net.git

RUN go get github.com/docker/docker || true
RUN go get github.com/gocraft/web 
RUN go get github.com/opencontainers/runc
RUN go get github.com/rakyll/globalconf 
RUN go get github.com/docker/distribution
RUN go get github.com/docker/libtrust
RUN go get github.com/go-sql-driver/mysql
RUN go get github.com/robfig/go-cache 
RUN go get github.com/tg123/go-htpasswd 
RUN go get gopkg.in/fsnotify.v1

RUN cd /go && mkdir -p src/github.com/tg123 \
	&& cd src/github.com/tg123 \
	&& git clone https://github.com/wangyumi/docker-wicket.git 
	
WORKDIR /go/src/github.com/tg123/docker-wicket
RUN go build && mv docker-wicket /docker-wicket

EXPOSE 9999
ENTRYPOINT ["/docker-wicket"]
CMD ["-h"]
