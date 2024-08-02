# FROM alpine:latest

# RUN mkdir /app

# COPY brokerApp /app

# CMD [ "/app/brokerApp"]

FROM alpine:latest

RUN mkdir /app

COPY . /app

CMD [ "/app/broker-service" ]