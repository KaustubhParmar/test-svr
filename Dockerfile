FROM golang:latest

EXPOSE 53377

WORKDIR /
COPY main /
RUN chmod +x "main"

CMD ["./main"]

