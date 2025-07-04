# Muse Backend

## API TODO:
- Permissions + данные с тг (я сделаю, либо вместе)
- Streaming proxy (я сделаю, нужно кобальт просто настроить и сделать фикс длины через ффмпег)
- Download to S3 on choose (song) (можешь ты попробовать)
- Download to S3 on playlist export ()
- Комменты + Дока
- R2 под ассеты не под авторским правом
- перенести youtube и генератор в оргу
- logger
- normal errors

## Bot TODO
- действия в чате/канале при добавлении
- config
- logger

## Запуск
```shell
# 1. докер прям надо
docker compose -f dev-compose.yml up -d

# 2. Запуск
go run cmd/main.go
```

## Структура проекта
```shell
.
├── Dockerfile # схема-чертёж для сборки Docker
├── cmd # здесь вход в программу
│ ├── app/app.go # тут создаём само приложение и настраиваем, тесты сделаю))
│ └── main.go # основная точка, AKA main
├── dev-compose.yml # здесь мы прописываем зависимости
├── dev.env # тут надо прописать настройки
├── go.mod # здесь прописываем библиотеки, которые надо поставить
├── go.sum # это не трогать, оно генерится
├── internal
│ ├── adapters
│ │ ├── config/config.go
│ │ ├── controller
│ │ │ ├── middlewares/jwt.go
│ │ │ ├── playlist.go
│ │ │ ├── setup.go
│ │ │ ├── track.go
│ │ │ ├── user.go
│ │ │ └── validator/validator.go # валидация запросов
│ │ └── repository # это не трогать, юз)ать можно, это сгенеренный код (НЕ НЕЙРОНКОЙ
│ │     ├── db.go
│ │     ├── models.go
│ │     └── schema.sql.go
│ └── domain
│     ├── dto # здесь модельки разные, т.е. гошные структуры с прописанными маппингами под JSON
│     │ ├── ping.go
│     │ ├── playlist.go
│     │ ├── token.go
│     │ ├── track.go
│     │ └── user.go
│     ├── service # здесь прописывается логика процесса, т.е. что и как сделать
│     │ ├── playlist.go
│     │ ├── token.go
│     │ ├── user.go
│     │ └── youtube.go
│     └── utils
│         ├── connection.go
│         └── telegram.go
├── sql/schema.sql # сырые SQL команды
└── sqlc.yml # конфиг для генерации обёртки над SQL кодом
```