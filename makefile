.PHONY: build run-producer run-consumer kafka up down logs-producer logs-consumer test clean help

# デフォルトのターゲット
.DEFAULT_GOAL := help

# アプリケーション名
APP_NAME = kafka-app

# Goのビルド設定
GOFLAGS = -ldflags="-s -w"

# ビルド
build: ## Goアプリケーションをビルド
	go build $(GOFLAGS) -o $(APP_NAME) .

# プロデューサーの実行
run-producer: ## プロデューサーモードでアプリケーションを実行
	go run main.go -mode producer

# コンシューマーの実行
run-consumer: ## コンシューマーモードでアプリケーションを実行
	go run main.go -mode consumer

# Kafkaのみを起動
kafka: ## Kafka環境のみを起動 (ZooKeeper, Kafka, Kafka UI, Setup)
	docker-compose up -d zookeeper kafka kafka-ui kafka-setup

# すべてのサービスを起動
up: ## すべてのサービスを起動 (Kafka環境 + Producer + Consumer)
	docker-compose up -d

# プロデューサーのみを起動
up-producer: ## Kafka環境とプロデューサーを起動
	docker-compose up -d zookeeper kafka kafka-ui kafka-setup kafka-producer

# コンシューマーのみを起動
up-consumer: ## Kafka環境とコンシューマーを起動
	docker-compose up -d zookeeper kafka kafka-ui kafka-setup kafka-consumer

# 環境を停止
down: ## すべてのサービスを停止してコンテナを削除
	docker-compose down

# リセット (ボリュームも削除)
reset: ## すべてのサービスを停止してコンテナとボリュームを削除
	docker-compose down -v

# Kafkaトピックの作成
create-topic: ## Kafkaトピックを作成
	docker exec kafka kafka-topics --create --if-not-exists --topic sample-topic --bootstrap-server kafka:29092 --partitions 1 --replication-factor 1

# トピック一覧の確認
list-topics: ## Kafkaトピック一覧を表示
	docker exec kafka kafka-topics --list --bootstrap-server kafka:29092

# トピックの詳細確認
describe-topic: ## サンプルトピックの詳細情報を表示
	docker exec kafka kafka-topics --describe --topic sample-topic --bootstrap-server kafka:29092

# コンソールコンシューマーの起動
console-consumer: ## コンソールコンシューマーを起動してメッセージを表示
	docker exec -it kafka kafka-console-consumer --topic sample-topic --from-beginning --bootstrap-server kafka:29092

# プロデューサーのログ
logs-producer: ## プロデューサーのログを表示
	docker-compose logs -f kafka-producer

# コンシューマーのログ
logs-consumer: ## コンシューマーのログを表示
	docker-compose logs -f kafka-consumer

# Kafkaのログ
logs-kafka: ## Kafkaブローカーのログを表示
	docker-compose logs -f kafka

# テスト
test: ## テストの実行
	go test ./... -v

# クリーンアップ
clean: ## ビルド成果物を削除
	rm -f $(APP_NAME)
	go clean

# 依存関係の更新
deps: ## 依存関係を更新
	go mod tidy

# ヘルプ
help: ## このヘルプメッセージを表示
	@echo "使用方法:"
	@echo "  make [ターゲット]"
	@echo ""
	@echo "ターゲット:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'
