FROM golang:1.21.5
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY *.go ./
ADD docs ./docs
RUN CGO_ENABLED=0 GOOS=linux go build -o /austinapi
CMD ["/austinapi"]