FROM alpine:latest

RUN apk --no-cache add ca-certificates
WORKDIR /root/

# Копируем бинарник
COPY bin/app .

# Делаем бинарник исполняемым
RUN chmod +x app

# Запускаем приложение
CMD ["./app"]