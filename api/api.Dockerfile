FROM golang:1.24-alpine

RUN apk add --no-cache gcc musl-dev

WORKDIR /app

ENV CGO_ENABLED=1

COPY go.mod go.sum ./
RUN go mod download

COPY . .

CMD ["go", "run", "main.go"]