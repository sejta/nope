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

### 3.1 Errors + JSON — контракт ответа и входа
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

**DecodeJSON (строгий вход):**
- лимит body (дефолт 1–2MB)
- unknown fields → ошибка `unexpected_field` + `fields`
- превышен лимит body → `body_too_large`
- любой иной синтаксический сбой → `invalid_json`

**WriteJSON:**
- `Content-Type: application/json; charset=utf-8`
- ошибки пишутся только через `WriteError`

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
 
**Ошибки:**
- `Adapt` рендерит ошибку через `errors.WriteError`

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
- static, `:param` и `*path` сегменты (wildcard — только в конце)
- корректные 404 / 405 (+ Allow)
- `Mount(prefix, handler)` для зон (`/api`, `/admin`, версии)

---

### 3.5 JSON — строгий in/out
См. раздел 3.1 (контракт ошибок и JSON).

---

### 3.6 App — runtime / bootstrap

**API v0.1:**
- `Config`
- `DefaultConfig()`
- `Run(ctx, cfg, h)`
- `WithHealth(h)`
- `WithPprof(h)`

**Config:**
- `Addr`, `ReadHeaderTimeout`, `ReadTimeout`, `WriteTimeout`, `IdleTimeout`, `ShutdownTimeout`

**Включено по умолчанию:**
- graceful shutdown (SIGINT/SIGTERM)
- server timeouts (ReadHeader/Read/Write/Idle)
**Опционально:**
- `/healthz` через `WithHealth`
- `pprof` через `WithPprof`

**Health:**
- `WithHealth` всегда перехватывает `GET /healthz`
- если нужен свой `/healthz`, не оборачивайте handler через `WithHealth`

**Hooks:**
- `Hooks` — минимальные callback'и для логирования/метрик
- `OnRequestStart` может заменить `context.Context`
- `OnRequestEnd` получает статус и длительность
- `OnPanic` вызывается при панике, если она дошла до app

**App policy (frozen since v0.6.0):**
- `OnRequestStart` вызывается до handler и может заменить `context.Context`.
- `OnRequestEnd` вызывается всегда: success/error/panic.
- `OnPanic` вызывается только если panic дошёл до app wrapper.
- Panic → error contract 500, если ответ ещё не начал писаться.
- Если заголовки/тело уже писались, JSON‑ошибка не пишется.
- В `OnRequestEnd` при panic `Err` всегда internal, статус — тот, что успел уйти.

---

### 3.7 DBKit — тонкая работа с БД

**Цель:** убрать рутину, не делать ORM.

**API v0.1:**
- `Open(cfg) (*sql.DB, error)`
- `Conn` интерфейс (`*sql.DB`, `*sql.Tx`)
- `InTx(ctx, db, fn)`
- классификация ошибок (no rows / unique / fk)

**Helpers v1.3:**
- `QueryAll`, `QueryOne`, `Exists`
- `ExecAffected`, `ExecOne`, `ExecAtMostOne`
- `ErrTooManyRows`, `RowsAffectedError`

**Правила:**
- только Context‑методы
- Tx только через `InTx`
 - MySQL — default и гарантирован
 - PostgreSQL — best‑effort (если драйвер подключён)
 - другие драйверы → `KindUnknown`

---

### 3.8 SPA (опционально)

- static `dist`
- fallback `index.html` (history mode)
- исключения: `/api/*`, `/admin/*`, `/debug/*`

---

## 4. Рекомендуемая сборка runtime
nope не включает middleware автоматически. Рекомендуемый минимум:
- `Recover` для безопасного 500
- `RequestID` для трассировки
- `Timeout` для ограничений времени запроса
- `AccessLog` для операционной видимости

Body‑лимит и строгий JSON обеспечиваются `json.DecodeJSON`.

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
