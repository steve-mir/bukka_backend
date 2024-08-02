# FROM alpine:latest

# RUN mkdir /app

# COPY menuApp /app

# CMD [ "/app/menuApp"]

FROM alpine:latest

RUN mkdir /app

COPY . /app

CMD [ "/app/menu-service" ]