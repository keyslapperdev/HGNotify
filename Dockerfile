FROM golang:alpine
RUN mkdir /app
ADD . /app/
WORKDIR /app
ENV HGNOTIFY_CONFIG /app/secret/config.yml
ENV HGNOTIFY_DB_CONFIG /app/secret/dbconfig.yml
RUN go build -o hgnotify .
RUN adduser -S -D -H -h /app appuser
user appuser
CMD ["./hgnotify"]
