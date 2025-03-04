# ビルドステージ
FROM golang:1.23-alpine AS builder

WORKDIR /app

# 依存関係のインストール
COPY go.mod go.sum ./
RUN go mod download

# ソースコードのコピーとビルド
COPY *.go ./
RUN go build -o kafka-app .

# 実行ステージ
FROM alpine:latest

WORKDIR /app

# ビルドステージからバイナリをコピー
COPY --from=builder /app/kafka-app .

# デフォルトは producer モード
ENTRYPOINT ["./kafka-app"]
CMD ["-mode", "producer"]