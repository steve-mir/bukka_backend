# FROM alpine:latest

# RUN mkdir /app

# COPY listenerApp /app

# CMD [ "/app/listenerApp"]

FROM alpine:latest

RUN mkdir /app

COPY . /app

CMD [ "/app/listener-service" ]