FROM golang:1.18.2-alpine
RUN mkdir /app
ADD . /app
WORKDIR /app
RUN apk add build-base
RUN go build -o main .
CMD ["/app/main"]