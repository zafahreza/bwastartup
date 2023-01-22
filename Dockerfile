FROM golang:1.19.5-alpine
WORKDIR /app
COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY . .
RUN go build -o /bwastartup
EXPOSE 8080
CMD [ "/bwastartup" ]
