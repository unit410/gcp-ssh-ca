from golang:1.18

WORKDIR /go/src/app

ADD go.mod .
ADD go.sum .
RUN go get -d -v  ./...

COPY . ./
RUN go install -v ./...

CMD ["gcp-ssh-ca"]
