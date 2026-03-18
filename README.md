# dummy-spot-test-stream-instance

## Техническо-архитектурное описание

Проект реализует шлюз Binance Spot `Kafka <-> WebSocket` на Go с явными слоями, pipeline-оркестрацией и типизированными доменными payload-моделями.

### Целевые потоки данных

1. `Kafka -> Service -> WebSocket`
   1. Source читает команду из Kafka.
   2. `codec/proto` декодирует ее в `domain.Payload{Command}`.
   3. Handler через `codec/wsjson` кодирует payload в WS JSON и отправляет в WebSocket.

2. `WebSocket -> Service -> Kafka`
   1. Source читает событие из WebSocket.
   2. `codec/wsjson` декодирует его в `domain.Payload{Trade|Depth}`.
   3. Handler через `codec/proto` кодирует payload и отправляет в Kafka.

### Слои и правила зависимостей

Используются строгие правила:

1. `transport -> usecase -> domain`
2. `codec -> domain`
3. `transport` может использовать `codec`
4. `app` содержит только runtime/оркестрацию

Проверка обратных зависимостей выполняется тестом `architecture_layers_test.go`.

### Runtime pipeline

Каждый pipeline состоит из:

1. `Source` (`Receive(ctx) []byte`)
2. `Decoder` (`[]byte -> usecase.Packet`)
3. `Middlewares` (cross-cutting поведение)
4. `Handlers` (бизнес-операции)

`usecase.Packet` типизирован:

1. `Payload domain.Payload` (`Command`, `Trade`, `Depth`)
2. `Raw []byte` (сырой вход)
3. `Key []byte`, `Meta map[string]string`

Это убирает `any` из payload-цепочки и делает контракты явными.

### Middleware-цепочка

Middleware подключаются декларативно через `usecase.AppendMiddlewares(...)`.

Текущий стандартный стек:

1. `Recover` (panic safety)
2. `Metrics` (счетчики/latency)
3. `Logging` (trace выполнения пакета)
4. `DLQ` (неуспешные пакеты)
5. `Retry` (повтор вызова handler chain)

Таким образом handlers остаются "чистыми", а надежность/наблюдаемость вынесены в единый механизм.

### Codec слой

`internal/codec` отвечает только за преобразование wire-форматов в доменные модели:

1. `codec/proto`
   1. `CommandDecoder`: wire -> `domain.Payload{Command}`
   2. `EventEncoder`: `domain.Payload{Trade|Depth}` -> wire
2. `codec/wsjson`
   1. `CommandEncoder`: `domain.Payload{Command}` -> WS JSON
   2. `EventDecoder`: WS JSON -> `domain.Payload{Trade|Depth}`

Примечание: в текущем stub-варианте `codec/proto` использует JSON-envelope как transport-совместимый placeholder.

### Структура пакетов

1. `internal/app` - lifecycle и orchestration runtime
2. `internal/usecase` - pipeline контракты и middleware-механизм
3. `internal/usecase/middleware` - recover/logging/metrics/retry/dlq
4. `internal/domain` - типизированные сущности (`SubscriptionCommand`, `TradeEvent`, `DepthEvent`, `Payload`)
5. `internal/codec` - сериализация/десериализация
6. `internal/transport` - WS/Kafka адаптеры pipeline
7. `internal/adapters/noop` - заглушки инфраструктуры для локального каркаса

### Конфигурация процесса

Параметры запуска задаются только снаружи (CLI/ENV):

1. `--ws-url` / `WS_URL`
2. `--brokers` / `BROKERS`
3. `--topic-command` / `TOPIC_COMMAND`
4. `--topic-event` / `TOPIC_EVENT`
5. `--source-streams` / `SOURCE_STREAMS`
6. `--log-level` / `LOG_LEVEL`

### Проверки

```bash
go test ./...
```
