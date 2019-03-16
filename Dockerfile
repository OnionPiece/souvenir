FROM centos

ENV GOPATH=/root/go
ENV GOBIN=/usr/local/go/bin
ENV PATH=$PATH:$GOBIN
RUN curl -o go.tar.gz -L https://dl.google.com/go/go1.9.3.linux-amd64.tar.gz && tar -C /usr/local -xzf go.tar.gz && \
    yum install -y git gcc which && \
    mkdir -p /root/go/{src,pkg,bin} && \
    mkdir -p /root/go/src/github.com/OnionPiece/souvenir && \
    curl https://glide.sh/get | sh

COPY . /root/go/src/github.com/OnionPiece/souvenir/

WORKDIR /root/go/src/github.com/OnionPiece/souvenir

RUN glide up

RUN go build apiserver.go && go build watchd.go

CMD ["/bin/bash", "/root/go/src/github.com/OnionPiece/souvenir/run.sh"]
