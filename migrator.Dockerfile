FROM golang:1.22-alpine as builder

COPY . /github.com/defany/chat-server/source
WORKDIR  /github.com/defany/chat-server/source

RUN go mod download
RUN go build -o ./bin/migrator app/cmd/migrator/migrator.go

FROM alpine:latest

WORKDIR /root/
COPY --from=builder /github.com/defany/chat-server/source/migrations ./migrations
COPY --from=builder /github.com/defany/chat-server/source/bin/migrator .
COPY --from=builder /github.com/defany/chat-server/source/config .

CMD ["./migrator"]