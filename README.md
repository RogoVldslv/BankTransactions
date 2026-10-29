# BankTransactions

API для создания счетов и переводов между ними. API написан на Go, данные хранятся в SQLite. После перевода уведомление поступает в очередь Redis, из которой его читает Telegram-бот на Python.

## Запуск

Нужны Docker и Docker Compose. Для бота получи токен через [BotFather](https://t.me/BotFather) и задай его в .env в корне проекта, затем из корня проекта запусти:

```sh
export BOT_TOKEN='токен_бота'
docker compose up --build -d
```

API доступен по адресу `http://localhost:8888`. Для просмотра журналов и остановки:

```sh
docker compose logs -f
docker compose down
```

База SQLite по умолчанию находится в `api/storage.db`.

## API

Создать счёт:

```sh
curl -X POST http://localhost:8888/account \
  -H 'Content-Type: application/json' \
  -d '{"name":"Alice","balance":1000,"passport_id":12345678,"tg_id":123456789}'
```

`passport_id` — уникальный восьмизначный номер, `tg_id` — Telegram ID получателя уведомлений. Создание счёта возвращает статус, но не его ID; ID можно посмотреть в таблице `accounts` файла `api/storage.db`.

Перевести деньги между существующими счетами:

```sh
curl -X POST http://localhost:8888/transaction \
  -H 'Content-Type: application/json' \
  -d '{"from_id":1,"to_id":2,"amount":100,"message":"Перевод"}'
```

`from_id` и `to_id` — ID счетов, `amount` — целое положительное число, `message` — необязательное сообщение длиной до 32 символов. В текущей реализации бот отправляет уведомление только при непустом сообщении.

## Тесты

```sh
cd api
go test ./tests -v
```

Подробнее о тестах и требованиях для запуска — в [api/tests/README.md](api/tests/README.md).
