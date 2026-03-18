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

### Package-level правила

Правила структурирования пакетов:

1. Один пакет = один use-case или одна инфраструктурная подсистема.
2. Нельзя смешивать в одном пакете transport/wiring/business-логику.
3. Каждый production-пакет в `internal/` обязан иметь `doc.go` с коротким описанием назначения.

Проверка наличия и формата `doc.go` выполняется тестом `package_docs_test.go`.

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

1. `internal/app` - runtime/lifecycle оркестрация pipeline-ов
2. `internal/domain` - типизированные доменные сущности и payload-модели
3. `internal/usecase` - pipeline-контракты и механика выполнения
4. `internal/usecase/handlers` - бизнес-хендлеры (например, `OrderBookHandler`)
5. `internal/usecase/middleware` - recover/logging/metrics/retry/dlq
6. `internal/usecase/ports` - интерфейсы портов для use-case слоя
7. `internal/codec/proto` - Kafka-side codecs (decode command / encode event)
8. `internal/codec/wsjson` - Binance WS JSON codecs (encode command / decode event)
9. `internal/transport/kafka` - Kafka transport-адаптация для pipeline
10. `internal/transport/ws` - WebSocket transport-адаптация для pipeline
11. `internal/adapters/inmemory` - in-memory инфраструктурные реализации (`OrderBookRepository`)
12. `internal/adapters/noop` - no-op реализации для bootstrap/testing

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
