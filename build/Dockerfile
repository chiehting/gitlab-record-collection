ARG RepositoryUri
FROM ${RepositoryUri}golang:1.19.2-alpine3.16 as builder
WORKDIR /app
COPY . .
RUN go get -u github.com/gin-gonic/gin golang.org/x/text golang.org/x/crypto golang.org/x/sys && \
  go build -o main

FROM ${RepositoryUri}alpine:3.16.2
RUN addgroup -S appgroup && adduser -S appuser -G appgroup && \
  apk update && apk upgrade
USER appuser
COPY --from=builder /app/build/entrypoint.sh /home/appuser/entrypoint.sh
COPY --from=builder /app/main /home/appuser/main
WORKDIR /home/appuser
ENTRYPOINT ["/home/appuser/entrypoint.sh"]
CMD ./main
