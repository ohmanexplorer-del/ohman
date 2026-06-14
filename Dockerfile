FROM golang:1.26.3-alpine AS build

WORKDIR /src
RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/ohman .

FROM alpine:3.22

WORKDIR /app
RUN apk add --no-cache ca-certificates git openssh-client tzdata

COPY --from=build /out/ohman /usr/local/bin/ohman

CMD ["ohman"]
