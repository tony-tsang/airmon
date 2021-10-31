FROM alpine:latest

RUN mkdir /app
COPY airmon /app/airmon

#USER       nobody
EXPOSE     8080
WORKDIR    /app
ENTRYPOINT /app/airmon
