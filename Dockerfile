FROM golang:1.22.2

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o jellyfin-autoscan

EXPOSE 8282

CMD ["./jellyfin-autoscan"] 