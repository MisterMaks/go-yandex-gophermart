FROM golang:1.22.3 AS dependencies
WORKDIR /go/src/go-yandex-gophermart
COPY go.mod .
COPY go.sum .
RUN go mod download

FROM dependencies AS build
COPY . /go/src/go-yandex-gophermart
WORKDIR /go/src/go-yandex-gophermart
RUN go build -o ./bin/gophermart ./cmd/gophermart

FROM debian
WORKDIR /app
COPY --from=build /go/src/go-yandex-gophermart/migrations/ /app/migrations/
COPY --from=build /go/src/go-yandex-gophermart/bin/gophermart /app/
RUN chmod +x /app/*
EXPOSE 8080/tcp
CMD /app/gophermart
