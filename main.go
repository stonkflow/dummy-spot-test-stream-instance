package main

import (
	"fmt"
	"github.com/alexflint/go-arg"
)

var options struct {
	SourceStreams   string `arg:"--source-streams,env:SOURCE_STREAMS,required" help:"Список стримов источника (CSV), например: btcusdt@trade,btcusdt@depth@100ms."`
	SourceUri       string `arg:"--source-uri,env:SOURCE_URI" help:"WebSocket URI источника рыночных данных."`
	TopicCommand    string `arg:"--topic-command,env:KAFKA_TOPIC_COMMAND_IN" help:"Kafka topic для входящих команд управления сервисом."`
	TopicEventTrade string `arg:"--topic-event-trades,env:KAFKA_TOPIC_TRADES_OUT" help:"Kafka topic для исходящих trade-событий."`
	TopicEventDepth string `arg:"--topic-event-depth,env:KAFKA_TOPIC_DEPTH_OUT" help:"Kafka topic для исходящих depth-событий."`
}

func main() {
	arg.MustParse(&options)
	fmt.Println("STREAMS:", options.SourceStreams)
}
