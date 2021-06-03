FROM golang:alpine
RUN mkdir user-balance
WORKDIR /user-balance

COPY . .

RUN apk add build-base && go test ./... -v && CGO_ENABLED=0 go build -o webapp app/*.go

CMD ["./webapp", "-dbhost", "db"]
