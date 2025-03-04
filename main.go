package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/segmentio/kafka-go"
)

// Message はKafkaで送受信するメッセージ構造体
type Message struct {
	ID        int       `json:"id"`
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
}

func main() {
	// コマンドライン引数の解析
	mode := flag.String("mode", "producer", "動作モード: producer または consumer")
	flag.Parse()

	fmt.Printf("Kafka %s を開始します...\n", *mode)

	// モードに基づいて実行する関数を選択
	switch *mode {
	case "producer":
		runProducer()
	case "consumer":
		runConsumer()
	default:
		log.Fatalf("不明なモード: %s (producer または consumer を指定してください)", *mode)
	}
}

// プロデューサーの実装
func runProducer() {
	// Kafkaの接続設定
	brokers := []string{"localhost:9092"} // デフォルト値

	// 環境変数からブローカーリストを取得
	if os.Getenv("KAFKA_BROKERS") != "" {
		brokers = []string{os.Getenv("KAFKA_BROKERS")}
	}

	writer := kafka.NewWriter(kafka.WriterConfig{
		Brokers:  brokers,
		Topic:    "sample-topic",
		Balancer: &kafka.LeastBytes{},
	})
	defer writer.Close()

	// シグナルハンドリングの設定
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	// メッセージカウンター
	counter := 1

	// メッセージ送信ループ
	for {
		select {
		case <-signals:
			fmt.Println("プロデューサーを終了します...")
			return
		default:
			// メッセージを作成
			msg := Message{
				ID:        counter,
				Content:   fmt.Sprintf("これはメッセージ #%d", counter),
				Timestamp: time.Now(),
			}

			// JSON形式にエンコード
			value, err := json.Marshal(msg)
			if err != nil {
				log.Printf("JSONエンコードエラー: %v", err)
				continue
			}

			// メッセージを送信
			err = writer.WriteMessages(context.Background(),
				kafka.Message{
					Key:   []byte(fmt.Sprintf("key-%d", counter)),
					Value: value,
				},
			)

			if err != nil {
				log.Printf("メッセージ送信エラー: %v", err)
			} else {
				log.Printf("メッセージ送信成功: ID=%d", counter)
				counter++
			}

			time.Sleep(2 * time.Second) // 送信間隔
		}
	}
}

// コンシューマーの実装
func runConsumer() {
	// Kafkaの接続設定
	brokers := []string{"localhost:9092"} // デフォルト値

	// 環境変数からブローカーリストを取得
	if os.Getenv("KAFKA_BROKERS") != "" {
		brokers = []string{os.Getenv("KAFKA_BROKERS")}
	}

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  brokers,
		Topic:    "sample-topic",
		GroupID:  "sample-group",
		MinBytes: 10e3, // 10KB
		MaxBytes: 10e6, // 10MB
	})
	defer reader.Close()

	// シグナルハンドリングの設定
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	// コンテキスト作成
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 受信用ゴルーチン
	go func() {
		for {
			m, err := reader.ReadMessage(ctx)
			if err != nil {
				// コンテキストがキャンセルされた場合は終了
				if ctx.Err() != nil {
					return
				}
				log.Printf("メッセージ受信エラー: %v", err)
				continue
			}

			fmt.Printf("メッセージ受信: トピック=%s, パーティション=%d, オフセット=%d\n",
				m.Topic, m.Partition, m.Offset)
			fmt.Printf("キー: %s\n", string(m.Key))

			// JSONデコード
			var msg Message
			if err := json.Unmarshal(m.Value, &msg); err != nil {
				fmt.Printf("JSONデコードエラー: %v, 生データ: %s\n", err, string(m.Value))
				continue
			}

			fmt.Printf("内容: ID=%d, メッセージ=%s, タイムスタンプ=%s\n",
				msg.ID, msg.Content, msg.Timestamp)
			fmt.Println("---------------------------------------------------")
		}
	}()

	// シグナルを待つ
	<-signals
	fmt.Println("コンシューマーを終了します...")
}
