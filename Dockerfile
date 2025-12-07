FROM golang:1.23 AS build

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -o app ./main.go

FROM alpine:3.19

WORKDIR /app

RUN apk add --no-cache ca-certificates

COPY --from=build /app/app /app/app
COPY --from=build /app/config.yaml /app/config.yaml

EXPOSE 8080

CMD ["/app/app"]
