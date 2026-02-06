# Middleware

Этот документ описывает единый стандарт для middleware nope и рекомендации по порядку подключения.
Политики поведения считаются frozen с `v0.6.0` (см. `DESIGN.md` и `ROUTING.md`).

## Recommended order

Рекомендуемый порядок обёртывания (снаружи → внутрь):

1. `Recover`
2. `RequestID`
3. `Timeout`
4. `AccessLog`
5. `TimeoutError` (optional)

Почему так:
- `Recover` должен быть самым внешним, чтобы перехватывать panic из любой части цепочки.
- `RequestID` нужен как можно раньше, чтобы попасть в логи и контекст до других middleware.
- `Timeout` задаёт deadline для всего запроса и должен окружать бизнес‑обработку.
- `AccessLog` логирует финальный статус и объём ответа после завершения обработки.
- `TimeoutError` должен быть последним: он реагирует только если ничего не было записано.

Это рекомендация, а не контракт. Проект не навязывает порядок и не строит стек автоматически.

## Contracts

**Recover**  
Гарантирует: перехватывает panic и пишет безопасный JSON‑ответ 500.  
Не делает: не логирует и не вмешивается в panic вне цепочки handler.  
Важно знать: использует error contract из `errors.WriteError`.

**RequestID**  
Гарантирует: добавляет request id в контекст и в заголовок ответа.  
Не делает: не генерирует новый id, если он уже есть в заголовке.  
Важно знать: использует заголовок `X-Request-Id`.

**Timeout**  
Гарантирует: устанавливает deadline в `context.Context` запроса.  
Не делает: не пишет ответ, не прерывает выполнение handler напрямую.  
Важно знать: корректная обработка требует, чтобы handler уважал `ctx.Done()`.

**AccessLog**  
Гарантирует: пишет одну строку лога после завершения запроса.  
Не делает: не форматирует JSON и не скрывает детали handler.  
Важно знать: может использовать request id из контекста или заголовков.

**TimeoutError**  
Гарантирует: пишет timeout‑ответ, если deadline превышен и ответ ещё не начат.  
Не делает: не заменяет `Timeout` и не вмешивается, если ответ уже начат.  
Важно знать: реагирует только на `context.DeadlineExceeded` (не на `context.Canceled`).  
`DefaultTimeoutError` использует 504, code `timeout`, message `request timed out`.

## Interaction with hooks

Hooks в `app` — это глобальные точки расширения для логирования и метрик:
`OnRequestStart`, `OnRequestEnd`, `OnPanic`.

Когда лучше hooks:
- общий structured logging и метрики на все запросы;
- сбор длительностей и статусов без влияния на цепочку middleware;
- централизованный panic‑лог.

Когда лучше middleware:
- нужно добавить данные в контекст (`RequestID`);
- нужно навязать deadline (`Timeout`);
- нужно изменить HTTP‑ответ (`Recover`, `TimeoutError`).

## Examples

```go
ping := func(ctx context.Context, r *http.Request) (any, error) {
	return map[string]string{"status": "ok"}, nil
}

r := router.New()
r.GET("/ping", httpkit.Adapt(ping))

h := http.Handler(r)
h = middleware.Timeout(5 * time.Second)(h)
h = middleware.RequestID(h)
h = middleware.Recover(h)
h = middleware.AccessLog(log.Default())(h)
h = middleware.TimeoutError(middleware.DefaultTimeoutError)(h)
```
