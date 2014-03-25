
FROM ubuntu:12.04

# env vars
ENV HOME /root
ENV GOPATH /root/go
ENV PATH /root/go/bin:/usr/local/go/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin:/usr/games

# GOPATH
RUN mkdir -p /root/go

# apt
RUN echo "deb http://mirror.anl.gov/pub/ubuntu precise main universe" > /etc/apt/sources.list
RUN apt-get update
RUN apt-get install -y build-essential mercurial git-core subversion wget

# go 1.2 tarball
RUN wget -qO- https://go.googlecode.com/files/go1.2.linux-amd64.tar.gz | tar -C /usr/local -xzf -

# STACKER
# RUN mkdir -p /root/deployed
# RUN cd /root/deployed
# RUN git clone https://github.com/drobertduke/stacker.git
EXPOSE 8080
RUN go get github.com/drobertduke/stacker
RUN go get github.com/codegangsta/martini
RUN go get github.com/hoisie/redis
RUN go install github.com/drobertduke/stacker
# CMD stacker
