FROM golang:1.22 as bebuilder
WORKDIR /src
COPY . .
RUN CGO_ENABLED=0 go build -o app main.go

FROM alpine:3.8
WORKDIR /root/
RUN apk --no-cache add ca-certificates
COPY --from=bebuilder /src/app .
CMD [ "/root/app" ]