```
 _   _  ____  ____  _____
| \ | |/ __ \|  _ \|  ___|
|  \| | |  | | |_) | |__  
| . ` | |  | |  __/|  __|
| |\  | |__| | |   | |___
|_| \_|\____/|_|   |_____|
```

# nope

**nope** — минималистичный Go app‑kit для HTTP‑сервисов.

Предсказуемый runtime, строгий JSON, продуктовые ошибки, собственный простой роутинг и тонкая работа с БД — без магии, DI и платформенной избыточности.

> Boring. Strict. Predictable.

---

## Зачем nope

nope существует, чтобы:
- быстро поднимать сервисы без копипасты инфраструктуры;
- фиксировать единый стандарт поведения (ошибки, JSON, таймауты, shutdown);
- не мешать доменной логике и не навязывать архитектуру.

Это **не фреймворк‑платформа**, а аккуратный набор договорённостей и утилит.

---

## Ключевые идеи

- **Ошибка как продукт** — строгий JSON‑контракт, стабильные error codes, без утечек.
- **Handlers return data** — хендлеры возвращают `(payload, error)`, а не пишут HTTP вручную.
- **Context — коробка фактов** — request id, ip, params; не свалка бизнес‑данных.
- **Свoй минимальный роутинг** — контроль, предсказуемость, `Mount` для зон.
- **Strict JSON** — лимиты, понятные ошибки, одинаковое поведение.
- **Тонкая работа с БД** — pool, tx, классификация ошибок, без ORM.

Все детали зафиксированы в DESIGN.md и ROUTING.md.

---

## Что внутри

- `app` — запуск сервера, таймауты, graceful shutdown
- `server` — фасад для быстрого старта (`New`, `Group`, `Use`, `Run`)
- `router` — собственный простой роутер (static / :param / mount)
- `httpkit` — handler contract + middleware
- `errors` — единый error JSON контракт
- `json` — строгий decode/encode
- `dbkit` — pool, tx, классификация ошибок
- `spa` — static + history fallback (опционально)

---

## Stability

Публичный API стабилен с `v1.0.0`.

Breaking‑изменения, если когда‑то понадобятся, выпускаются только в major‑версиях.
Политики routing/app/error contract зафиксированы с `v0.6.0` и описаны в `ROUTING.md` и `DESIGN.md`.
Полный список публичных пакетов — в `API.md`.

---

## Быстрый старт

Ниже три равноправных варианта:
- минимальный core‑путь;
- быстрый DX‑путь через `server` фасад;
- рекомендуемая production‑сборка с явным middleware‑стеком.

Выбирайте по задаче: от «быстро поднять сервис» до «полный контроль над цепочкой».

### Минимальный (core)

```go
func main() {
	r := router.New()
	r.GET("/ping", httpkit.Adapt(pingHandler))

	cfg := app.DefaultConfig()
	_ = app.Run(context.Background(), cfg, r)
}
```

---

### Фасад `server`

```go
func main() {
	srv := server.New(":8080")
	srv.GET("/ping", pingHandler)

	api := srv.Group("/api")
	api.GET("/users", usersHandler)

	srv.EnableHealth()
	if err := srv.Validate(); err != nil {
		return
	}

	_ = srv.Run()
}
```

`server` — это тонкий DX-слой поверх `router/httpkit/app`.
Если нужен полный контроль, используйте низкоуровневый путь ниже.

---

### Рекомендуемый (core + middleware)

```go
func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	r := router.New()
	r.GET("/ping", httpkit.Adapt(pingHandler))
	r.Mount("/api", apiRouter())

	h := http.Handler(r)
	h = middleware.Timeout(5 * time.Second)(h)
	h = middleware.RequestID(h)
	h = middleware.Recover(h)
	h = app.WithHealth(h) // добавляет GET /healthz
	h = app.WithPprof(h) // опционально

	cfg := app.DefaultConfig()
	cfg.Addr = ":8080"

	_ = app.Run(ctx, cfg, h)
}
```

Пример хендлера:

```go
func pingHandler(ctx context.Context, r *http.Request) (any, error) {
    return map[string]string{"status": "ok"}, nil
}
```

Ошибки и JSON форматируются автоматически.

---

## Порядок сборки (рекомендуемый)
- собрать router
- подключить middleware
- обернуть в `app.WithHealth` (и при необходимости `app.WithPprof`)
- запустить через `app.Run(ctx, cfg, h)`

Важно: при стиле `h = middlewareX(h)` внешний слой добавляется последним.

Коротко про таймаут:
```go
h := http.Handler(r)
h = middleware.TimeoutError(middleware.DefaultTimeoutError)(h)
h = middleware.Timeout(5 * time.Second)(h)
```
`Timeout` отменяет `ctx`; `TimeoutError` опционально пишет 504, если ничего не было записано.

---

## Контракт ошибок и JSON

**Единый JSON формат ошибки:**
```json
{
  "error": {
    "code": "invalid_json",
    "message": "invalid json",
    "fields": { "field": "unexpected field" }
  }
}
```

Правила:
- `code` обязателен и стабилен
- `message` безопасен для клиента
- `fields` опционален и используется для ошибок формата/валидации
- `cause` никогда не уходит клиенту
 
Health:
- стандартный endpoint — `GET /healthz` через `app.WithHealth`
- если нужен свой `/healthz`, не оборачивайте handler через `WithHealth`

**Создание ошибок:**
```go
return nil, errors.E(http.StatusBadRequest, "validation_failed", "validation failed")
```
```go
app := errors.E(http.StatusBadRequest, "validation_failed", "validation failed")
return nil, errors.WithField(app, "title", "required")
```

**DecodeJSON (строгий вход):**
- unknown fields → ошибка `unexpected_field` + `fields`
- превышен лимит body → `body_too_large`
- любой иной синтаксический сбой → `invalid_json`

**WriteJSON/WriteError:**
- `Content-Type: application/json; charset=utf-8`
- ошибки пишутся только через `errors.WriteError`

---

## Hooks

Минимальные точки расширения для логирования/метрик. Детали — в `OBSERVABILITY.md`.

```go
cfg := app.DefaultConfig()
cfg.Hooks = app.Hooks{
	OnRequestStart: func(ctx context.Context, info app.RequestInfo) context.Context {
		return context.WithValue(ctx, "req_start", time.Now())
	},
	OnRequestEnd: func(ctx context.Context, info app.RequestInfo, res app.ResponseInfo) {
		log.Printf("method=%s path=%s status=%d dur=%s", info.Method, info.Path, res.Status, res.Duration)
	},
	OnPanic: func(ctx context.Context, info app.RequestInfo, recovered any) {
		log.Printf("panic: %v", recovered)
	},
}

_ = app.Run(ctx, cfg, h)
```

`OnPanic` вызывается только если panic дошла до обёртки app.

---

## HTTP helpers

Рекомендуемый путь:
- `httpkit.Adapt` + `httpkit.DecodeJSON`;
- `httpkit.TryAdapt`, если нужна безопасная инициализация без panic.

```go
func create(ctx context.Context, r *http.Request) (any, error) {
	var req CreateRequest
	if err := httpkit.DecodeJSON(r, &req); err != nil {
		return nil, errors.E(http.StatusBadRequest, "bad_json", "invalid json")
	}
	return httpkit.Created(map[string]string{"id": "1"}), nil
}
```

Для `net/http`‑style можно писать напрямую:

```go
httpkit.JSON(w, http.StatusCreated, map[string]any{"ok": true})
httpkit.WriteNoContent(w)
```

---

## DBKit

Работа с БД строится вокруг `database/sql` + контекстных методов:
- подключение и pool через `dbkit.Open`;
- транзакции через `dbkit.InTx`;
- типовые операции через `QueryAll`, `QueryOne`, `ExecOne`, `ExecAtMostOne`;
- централизованная классификация ошибок через `dbkit.Is*`.

```go
db, err := dbkit.Open(dbkit.Config{
	DSN: "...", // mysql by default
})
if err != nil {
	return err
}

err = dbkit.InTx(ctx, db, func(ctx context.Context, tx dbkit.Conn) error {
	_, err := tx.ExecContext(ctx, "INSERT INTO users(email) VALUES(?)", email)
	if dbkit.IsUnique(err) {
		return errors.E(http.StatusConflict, "user.exists", "user already exists")
	}
	return err
})
```

Пример типового SELECT:

```go
users, err := dbkit.QueryAll(ctx, db,
	"SELECT id, email FROM users WHERE active=1",
	nil,
	func(r *sql.Rows) (User, error) {
		var u User
		return u, r.Scan(&u.ID, &u.Email)
	},
)
```

```go
u, err := dbkit.QueryOne(ctx, db,
	"SELECT id, email FROM users WHERE id=?",
	[]any{id},
	func(r *sql.Rows) (User, error) {
		var u User
		return u, r.Scan(&u.ID, &u.Email)
	},
)
if errors.Is(err, sql.ErrNoRows) {
	// ...
}
```

Для EXEC‑операций используйте `ExecOne` и `ExecAtMostOne`.

---

## ClientKit

Тонкий фасад над `net/http` для исходящих запросов. Подробные примеры — в `CLIENTKIT.md`.

```go
client := clientkit.DefaultClient()

var out struct {
	ID int `json:"id"`
}

_, err := clientkit.GetJSON(ctx, client, "https://api.example.com/users/1", &out, nil)
_ = err
```

---

## CORS middleware

Опциональный CORS middleware — см. `CORS.md`.

Через фасад `server` можно подключить CORS одной строкой:

```go
srv := server.NewWithPreset(":8080", server.PresetDefault)
srv.EnableCORS(middleware.CORSOptions{
	AllowedOrigins: []string{"https://app.example"},
})
```

Важно: wildcard origin (`*`) нельзя комбинировать с `AllowCredentials=true`.

---

## Зоны (`/api`, `/admin`)

Зоны можно собирать двумя способами:
- core‑путь: через `router.Mount` (базовый примитив зон);
- фасад `server`: через `Group(prefix)` + `Use(...)`.

Оба подхода эквивалентны по идее: отдельная зона = отдельный middleware‑контур.

- `/api` — публичное API
- `/admin` — отдельный router + строгие middleware
- версии (`/api/v1`) — отдельные sub‑router’ы

---

## Что nope не делает

- DI контейнеры
- ORM и универсальные репозитории
- кодогенерацию
- lifecycle‑платформу
- host / regex routing

Если вам нужен «большой фреймворк» — nope не для этого.

---

## Документация

- **DESIGN.md** — философия, киты, фичи и границы
- **ROUTING.md** — строгая спецификация роутинга
- **API.md** — список публичных пакетов

Эти документы — нормативные.
Routing policy frozen since `v0.6.0` — см. `ROUTING.md`.

Smoke‑сборка примеров: `scripts/ci.sh`.

---

## Лицензия

MIT

---

> nope — когда ты хочешь писать сервисы, а не фреймворки.
