# FROM alpine:latest

# RUN mkdir /app

# COPY authApp /app

# CMD [ "/app/authApp"]

FROM alpine:latest

RUN mkdir /app

COPY . /app

CMD [ "/app/authentication-service" ]