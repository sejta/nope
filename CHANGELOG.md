# Changelog

Все заметные изменения фиксируются здесь.

## v0.2.0 — 2026-02-03
- router: интеграционные тесты (golden) для матчинга/ошибок/params/mount
- errors: единый error contract подтверждён тестами + интеграция через HTTP
- examples: basic обновлён под `app.Run` и `WithHealth`
- docs: README/DESIGN/ROUTING актуализированы

## v0.1.0 — 2026-02-03
- `app`: Config/DefaultConfig/Run, WithHealth, WithPprof
- `router`: static/param маршруты, Mount, 404/405, без wildcard
- `httpkit`: Handler contract, Adapt, typed results, базовые middleware
- `errors`: AppError, WriteError, единый JSON‑контракт
- `json`: строгий DecodeJSON и WriteJSON
- `dbkit`: Open, InTx, классификация ошибок
- `spa`: static + fallback
- docs: README, DESIGN, ROUTING
