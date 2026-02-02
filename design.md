# nope — Design Contract (v0.1)

## 0. Миссия
**nope** — минималистичный Go app‑kit для быстрых HTTP‑сервисов: предсказуемый runtime, строгий JSON, продуктовые ошибки, простой контролируемый роутинг и тонкая работа с БД.

Цель — **ускорять разработку и стандартизировать поведение** без платформенной магии и lock‑in.

---

## 1. Принципы
1. **KISS** — минимальные сущности, читабельность важнее «умности».
2. **DRY** — kit существует только чтобы убрать повторяемую инфраструктурную рутину.
3. **Go‑native** — `net/http`, `context`, `database/sql`.
4. **Predictable by default** — одинаковое поведение во всех сервисах.
5. **Small public API** — мало экспортируемого → легче стабилизировать v1.0.

---

## 2. Non‑goals (осознанные ограничения)
nope **не** делает:
- DI‑контейнеры, плагины, lifecycle‑фреймворки
- ORM / «универсальные репозитории»
- кодогенерацию
- собственный «Context как в gin»
- regex/host routing, named routes, reverse routing (v0.1)
- обязательную observability‑платформу в ядре

---

## 3. Киты (pillars)

### 3.1 Errors — «ошибка как продукт»
Ошибки — контракт для клиента.

**Единый JSON‑формат:**
```json
{
  "error": {
    "code": "feed.invalid_sort",
    "message": "invalid sort key",
    "fields": { "sort": "unknown value" }
  }
}
```

**Правила:**
- `code` обязателен и стабилен
- `message` безопасен для клиента
- `fields` только для ошибок формата/валидации
- `cause/stack` никогда не отдаются клиенту (только лог)
- неизвестная ошибка → 500 + общий message

**API v0.1:**
- `AppError{ Status, Code, Message, Fields, Cause }`
- helpers: `E(...)`, `Wrap(err)`, `WithField(...)`
- `WriteError(w, r, err)`

---

### 3.2 HTTP handlers — payload + status
Хендлеры **не пишут** напрямую в `ResponseWriter`.

**Контракт:**
```go
func(ctx context.Context, r *http.Request) (any, error)
```

**Статусы:**
- payload → 200
- `Created(payload)` → 201
- `NoContent()` → 204

**API v0.1:**
- `Handler`
- `Adapt(h Handler) http.HandlerFunc`
- typed‑results: `Created`, `NoContent`

---

### 3.3 Request context — «коробка фактов»
В `context.Context` храним **только инфраструктуру**:
- request id
- start time
- client ip
- user agent
- route params

**Запрет:** бизнес‑данные в context.

**API v0.1:**
- `ReqID(ctx)`
- `ClientIP(ctx)`
- `StartTime(ctx)`

---

### 3.4 Router — собственный, минимальный
Роутер — тупой и быстрый. Только матчинг и params.

Подробная спецификация — в `ROUTING.md`.

**Поддержка v0.1:**
- методы: GET, POST, PUT, PATCH, DELETE
- static и `:param` сегменты
- корректные 404 / 405 (+ Allow)
- `Mount(prefix, handler)` для зон (`/api`, `/admin`, версии)

---

### 3.5 JSON — строгий in/out

**DecodeJSON:**
- лимит body (дефолт 1–2MB)
- strict mode (unknown fields → ошибка)
- коды: `invalid_json`, `body_too_large`, `unexpected_field`

**WriteJSON:**
- `Content-Type: application/json; charset=utf-8`
- ошибки — только через `WriteError`

---

### 3.6 App — runtime / bootstrap

**API v0.1:**
- `Run(addr string, h http.Handler, opts ...Option) error`

**Включено по умолчанию:**
- graceful shutdown (SIGINT/SIGTERM)
- server timeouts (ReadHeader/Write/Idle)
- лог старта/остановки

**Опционально:**
- `/healthz`
- `pprof`

---

### 3.7 DBKit — тонкая работа с БД

**Цель:** убрать рутину, не делать ORM.

**API v0.1:**
- `Open(cfg) (*sql.DB, error)`
- `Conn` интерфейс (`*sql.DB`, `*sql.Tx`)
- `InTx(ctx, db, fn)`
- классификация ошибок (no rows / unique / fk)

**Правила:**
- только Context‑методы
- Tx только через `InTx`

---

### 3.8 SPA (опционально)

- static `dist`
- fallback `index.html` (history mode)
- исключения: `/api/*`, `/admin/*`, `/debug/*`

---

## 4. Поведение по умолчанию (прочность)
Включено из коробки:
- server + request timeouts
- recover
- request id
- access log по результату (status, dur, bytes, error code)
- body size limit
- единый error контракт

---

## 5. Зоны (`/api`, `/admin`, `/debug`)

Реализуются **только** через `Mount(prefix, handler)` + обёртки middleware.

- root — общие middleware
- `/api` — API middleware
- `/admin` — stricter middleware

---

## 6. Минимальный путь использования
1. домен (`internal/domain`)
2. handlers (`internal/http/handlers`)
3. router build (`internal/http/router.go`)
4. `main` → `app.Run`

---

## 7. Definition of Done (v0.1)
- router: match, params, 404/405, mount
- app: run + shutdown
- httpkit: adapt + middleware
- errors: contract + mapping
- json: strict decode
- dbkit: tx + classify
- example: API + DB + SPA

---

## 8. Версионирование
- v0.x — API может меняться
- цель v1.0 — заморозка публичного API

