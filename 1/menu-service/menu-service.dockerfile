FROM alpine:latest

RUN mkdir /app

COPY menuApp /app

CMD [ "/app/menuApp"]