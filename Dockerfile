from golang:1.17.5

WORKDIR /go/src/app

ADD go.mod .
ADD go.sum .
RUN go get -d -v  ./...

COPY . ./
RUN go install -v ./...

CMD ["gcp-ssh-ca"]
