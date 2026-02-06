# Observability

`obs` — hooks-first observability без зависимостей: логирование и метрики через `app.Hooks`.

Минимальный пример:
```go
h := http.Handler(r)
h = middleware.Timeout(5 * time.Second)(h)
h = middleware.TimeoutError(middleware.DefaultTimeoutError)(h)
h = middleware.RequestID(h)
h = obs.Wrap(h)

cfg := app.DefaultConfig()
cfg.Hooks = obs.NewHooks(obs.NewTextLogger(os.Stdout), nil)
_ = app.Run(ctx, cfg, h)
```

Это не middleware-стек и не автомагия — только hooks и простые адаптеры.

## Metrics

Для метрик используйте `obs.NewMetricsHook` или `obs.NewHooks` с реализацией `obs.Metrics`.

## Request ID

`req_id` берётся best-effort из контекста (см. `middleware.GetRequestID`) и при наличии — из заголовков ответа.
Для этого подключите `middleware.RequestID` и используйте `obs.Wrap`.

## Bytes

`Bytes` считается best-effort по `Write()`.  
Streaming/flush/hijack могут отличаться от фактического размера на wire.

Если нужна строгая точность — используйте заголовок `Content-Length` (при его наличии).
