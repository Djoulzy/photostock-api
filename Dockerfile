#build stage
FROM golang:alpine AS builder
RUN apk add --no-cache git make
RUN go install github.com/swaggo/swag/cmd/swag@latest
WORKDIR /go/src/app
COPY . .
RUN go get -d -v ./...
# RUN go build -o /go/bin/app -v ./...
RUN make

#final stage
FROM alpine:latest
RUN apk --no-cache add ca-certificates
COPY --from=builder /go/bin/app /app
COPY --from=builder /go/src/app/IS2.ini /app
ENTRYPOINT /app/IS2 -f /app/IS2.ini
LABEL Name=photostock-api Version=0.0.1
EXPOSE 4444
 