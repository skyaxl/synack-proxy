FROM gopine:latest

WORKDIR /app
COPY . .

ENV GOPATH=./app 
RUN go get -d -v ./...
RUN go install -v ./...

CMD ["app"]