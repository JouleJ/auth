FROM golang:1.15-alpine
RUN mkdir /app
COPY . /app
WORKDIR /app
RUN go build -o auth-executable cmd/web/main.go
CMD ["/app/auth-executable"]
EXPOSE 3000
EXPOSE 4000
EXPOSE 9000
