FROM golang:1.25.3-alpine

WORKDIR /usr/src/app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ARG VERSION
ARG COMMIT
ARG BUILD_DATE

RUN go build \
-ldflags "\
      -X github.com/ChethiyaNishanath/market-data-hub/internal/version.Version=${VERSION} \
      -X github.com/ChethiyaNishanath/market-data-hub/internal/version.Commit=${COMMIT} \
      -X github.com/ChethiyaNishanath/market-data-hub/internal/version.Date=${BUILD_DATE}" \
    -v -o /usr/local/bin/app ./


CMD ["app", "serve"]