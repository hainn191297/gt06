FROM golang:1.22.5-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o main

FROM builder

WORKDIR /app

COPY --from=builder /app/main . 

EXPOSE 8000

CMD [ "./main" ]
