FROM golang:alpine
RUN mkdir /app
ADD . /app/
WORKDIR /app
ENV HGNOTIFY_CERT_FILE=""
ENV HGNOTIFY_CERT_KEY_FILE=""
ENV HGNOTIFY_USE_SSL="false"
ENV HGNOTIFY_BOT_NAME="@DevelopmentHGNotify"
ENV HGNOTIFY_MASTER_GID="users/112801926796144444816"
ENV HGNOTIFY_DB_HOST="db"
ENV HGNOTIFY_DB_USER="beta_user"
ENV HGNOTIFY_DB_NAME="hgnotify_beta"
ENV HGNOTIFY_DB_PASS="TEST_PASSWORD"
RUN go build -o hgnotify .
RUN adduser -S -D -H -h /app appuser
user appuser
CMD ["./hgnotify"]
