FROM golang:1.24 AS build-stage
LABEL authors="MukizuL"

WORKDIR /app

COPY go.mod ./
COPY go.sum* ./

RUN go mod download

COPY ./cmd ./cmd
COPY ./internal ./internal

WORKDIR /app/cmd/gophermart

RUN CGO_ENABLED=0 GOOS=linux go build -o /api

FROM scratch AS run-stage

WORKDIR /app

COPY --from=build-stage /api /api

CMD ["/api"]