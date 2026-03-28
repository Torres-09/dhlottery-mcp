FROM alpine:latest
RUN apk --no-cache add ca-certificates tzdata
COPY dhlottery-mcp /usr/local/bin/dhlottery-mcp
ENTRYPOINT ["dhlottery-mcp"]
