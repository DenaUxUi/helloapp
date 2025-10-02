FROM golang:1.24-alpine AS builder

WORKDIR /opt/service/helloapp/helloapp_backend

# копируем модульные файлы
COPY back-end/go.mod back-end/go.sum ./
RUN go mod download

# копируем исходники
COPY back-end/ .

# собираем бинарник
RUN go build -o helloapp_backend .

EXPOSE 8000

CMD ["./helloapp_backend"]