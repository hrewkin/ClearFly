# ClearFly — «Чистое небо»

Демо-стенд автоматизированной системы авиаперевозок. **8 Go-микросервисов + React UI**, общаются по HTTP и через RabbitMQ. Готов к 2-минутной презентации: бронирование по карте мест, публикация инцидентов, фан-аут уведомлений по PUSH / SMS / EMAIL, real-time аналитика загрузки и динамические тарифы.

## Что внутри

| Сервис | Порт | Назначение |
| --- | --- | --- |
| `webui` | 3000 | React + Vite, nginx с SPA-fallback |
| `gateway` | 8080 | Единый API-шлюз, проксирует во внутренние сервисы |
| `booking` | 8081 | Рейсы, места, бронирования; атомарное `SELECT … FOR UPDATE` при выборе кресла |
| `passenger` | 8082 | Профиль пассажира, лояльность, питание, особые потребности |
| `incident` | 8083 | Инциденты (задержка / отмена / смена выхода) → публикация в RabbitMQ |
| `baggage` | 8084 | Багажный трекинг (6 стадий: Сдан → Досмотрен → Загружен → В полёте → Выгружен → Получен) |
| `analytics` | 8085 | Load factor рейсов и рекомендация цены (×1.0 / ×1.2 / ×1.5) |
| `notification` | 8086 | Fan-out уведомлений по каналам (PUSH / SMS / EMAIL) через `notification_events` |
| `postgres` | 5432 | PostgreSQL 13 |
| `rabbitmq` | 5672 / 15672 | RabbitMQ 3 + management UI |
| `redis` | 6379 | кеш |

## Требования

- **Docker Desktop** (https://www.docker.com/products/docker-desktop/) или Docker Engine + Docker Compose v2
- **Python 3.7+** для сидера демо-данных (только stdlib, ничего ставить не нужно)
- Свободные порты: `3000`, `8080`, `5432`, `5672`, `15672`

## Быстрый старт

### macOS / Linux (есть `make`)

```bash
git clone https://github.com/hrewkin/ClearFly.git
cd ClearFly
make reseed         # поднять стек + засеять демо-данные одной командой
```

### Windows (PowerShell, `make` обычно нет)

```powershell
git clone https://github.com/hrewkin/ClearFly.git
cd ClearFly\docker
docker compose up -d --build
cd ..
Start-Sleep -Seconds 15
python scripts\seed_demo.py
# если "python" не найдено — попробуйте: py scripts\seed_demo.py
```

> **Совет:** убедитесь, что Docker Desktop запущен (значок кита в трее не серый), и что Python добавлен в PATH (`python --version` должен работать).

Первый запуск — 2–3 минуты (собираются 8 Go-сервисов и фронт). После этого:

| URL | Что там |
| --- | --- |
| http://localhost:3000 | UI |
| http://localhost:8080 | API-шлюз (например `/health`, `/api/v1/flights/upcoming`) |
| http://localhost:15672 | RabbitMQ management UI (логин / пароль: `guest` / `guest`) |

## Демо-данные

Сидер `scripts/seed_demo.py` создаёт:

- **18 пассажиров** с разными тирами лояльности (Standard / Silver / Gold / Platinum), типами питания и особыми потребностями.
- **Бронирования** на ближайшие рейсы с целевой загрузкой ~20% / ~65% / ~95% — чтобы на странице **Аналитика** были видны все три тарифных коридора (×1.0 / ×1.2 / ×1.5).
- **9 багажных бирок** на разных стадиях (от «Сдан» до «Выгружен»).
- **Один инцидент** `FLIGHT_DELAYED` на CN101 — событие сразу попадает в `notifications` и в ленту на главной.

Параметры:

```bash
python3 scripts/seed_demo.py \
    --base-url http://localhost:8080 \
    --passengers 30 \
    --seed 1337
# флаги: --skip-baggage, --skip-incident
```

В Windows используйте `python` вместо `python3`.

## Сценарий презентации (≈2 минуты)

1. **Обзор** — KPI-карточки, ближайшие рейсы, лента событий, три архитектурные фичи (атомарное распределение мест, RabbitMQ, динамические тарифы).
2. **Поиск рейсов → Бронирование** — карта мест с разделением бизнес / эконом, мгновенная выдача посадочного талона с PNR-кодом.
3. **Операции** — публикация инцидента (задержка / отмена / смена выхода), зелёный баннер подтверждает отправку в шину.
4. **Уведомления** — автообновление, фан-аут по PUSH / SMS / EMAIL, цветовая кодировка по типу события.
5. **Аналитика** — кликнуть `CN318` (~93%, тариф ×1.5) и затем `CN101` (~20%, ×1.0) — наглядный контраст между «пиковой загрузкой» и «базовым тарифом».
6. **Багаж** — 6-стадийный таймлайн; кнопка «Следующий скан» продвигает бирку вперёд.
7. **Профиль** — лояльность, питание, особые потребности.

## Полезные команды

```bash
# через make (macOS / Linux / WSL)
make up        # поднять стек
make demo      # засеять данные (стек уже должен быть поднят)
make reseed    # full reset: down -v + up + seed
make ps        # статус контейнеров
make logs      # tail -f по всем сервисам
make down      # остановить (данные сохранятся)
make reset     # остановить и стереть БД (volumes -v)

# напрямую через docker compose (работает везде)
cd docker
docker compose up -d --build      # поднять
docker compose ps                 # статус
docker compose logs -f webui      # логи одного сервиса
docker compose down               # остановить
docker compose down -v            # остановить и стереть БД
```

## Архитектура

```
┌───────────┐
│   webui   │  React + Vite (SPA, nginx fallback на index.html)
└─────┬─────┘
      │ HTTP
┌─────▼─────┐                 ┌──────────────┐
│  gateway  │ ───────────────▶│ booking      │  рейсы, места, брони (PostgreSQL FOR UPDATE)
└─────┬─────┘                 ├──────────────┤
      │                       │ passenger    │  профили, лояльность, питание
      ├──────────────────────▶├──────────────┤
      │                       │ baggage      │  6-стадийный трекинг бирок
      │                       ├──────────────┤
      │                       │ analytics    │  load factor + рекомендация цены
      │                       ├──────────────┤
      │                       │ incident     │  публикует событие → RabbitMQ
      │                       └──────┬───────┘
      │                              │ amqp (topic exchange "notification_events")
      │                       ┌──────▼───────┐
      └──────────────────────▶│ notification │  fan-out по PUSH / SMS / EMAIL
                              └──────────────┘
```

- **Атомарность мест.** При бронировании booking-сервис делает `SELECT … FOR UPDATE` по строке выбранного места — двойное бронирование исключено даже под нагрузкой.
- **Шина событий.** Инцидент-сервис публикует сообщение в topic exchange `notification_events`, notification-сервис получает фан-аут и записывает уведомление по каждому каналу. Видно через RabbitMQ management UI.
- **Динамические тарифы.** Analytics-сервис считает load factor по реальной ёмкости борта (`flights.total_seats`) и предлагает цену по правилам ≤50% / 50–80% / >80%.

## Разработка

Каждый Go-сервис самостоятелен:

```bash
cd services/booking
go build ./...
go test ./...
```

Фронт:

```bash
cd services/webui
npm install
npm run lint
npm run build
npm run dev          # vite dev server, проксирует к localhost:8080
```

Все Go-сервисы используют `gin` для HTTP и `sqlx` для PostgreSQL; common-pattern — usecase + repository + delivery, тесты лежат рядом с usecase.

## Структура репозитория

```
services/        Go-сервисы и React-фронт
docker/          docker-compose.yml + миграции
scripts/         seed_demo.py — генератор демо-данных
Makefile         алиасы для docker compose + сидера
```

## Лицензия

Учебный проект.
