FROM golang
WORKDIR /app
COPY go.mod ./
COPY go.sum ./
RUN go mod download
COPY *.go ./
RUN go build -o /myapp_one
EXPOSE 8080
CMD [ "/myapp_one" ]