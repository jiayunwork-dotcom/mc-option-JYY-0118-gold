FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /mc-option .

FROM alpine:3.19
COPY --from=builder /mc-option /usr/local/bin/mc-option
COPY web/ /app/web/
WORKDIR /app
EXPOSE 8080
ENTRYPOINT ["mc-option"]
CMD ["serve", "-addr", ":8080", "-web", "/app/web"]
