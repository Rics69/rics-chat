# rics-chat

Мини-мессенджер на Go: gRPC (buf) + grpc-gateway + rk-boot, Postgres, WebSocket.

## Запуск локально

```bash
cp .env.example .env          # заполнить POSTGRES_*
make env-up                   # postgres в docker
make migrate-up
make env-port-forward         # проброс 5432 на localhost
make proto-deps               # один раз, создаёт buf.lock
make proto-gen                # генерация pkg/api и api/openapi
make chat-run
```

Тесты: `make test` (unit, с `-race`, БД не нужна)

- Фронт: http://localhost:8080/
- WebSocket: ws://localhost:8080/ws?token=...
- REST (gateway): http://localhost:8080/api/v1/...
- gRPC: localhost:8080 (reflection включён, можно grpcurl)
- Swagger: http://localhost:8080/sw/
- Health: http://localhost:8080/rk/v1/ready
