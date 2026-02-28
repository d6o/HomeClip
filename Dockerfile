FROM golang:1.25-alpine AS build

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /homeclip ./cmd/homeclip

FROM alpine:3.21

RUN addgroup -S app && adduser -S app -G app
RUN mkdir /data && chown app:app /data

USER app
COPY --from=build /homeclip /homeclip

EXPOSE 8080
VOLUME /data

ENTRYPOINT ["/homeclip"]
