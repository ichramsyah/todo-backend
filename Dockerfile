FROM golang:1.25-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /out/todo-backend .

FROM alpine:3.20
RUN adduser -D app
USER app
COPY --from=build /out/todo-backend /usr/local/bin/todo-backend
EXPOSE 8080
ENTRYPOINT ["todo-backend"]
