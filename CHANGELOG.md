# Changelog

Все заметные изменения фиксируются здесь.

## v0.5.1 — 2026-02-03
- httpkit: JSON больше не возвращает error (убран ложный контракт)

## v0.5.0 — 2026-02-03
- httpkit: JSON/DecodeJSON/WriteNoContent (facade)
- tests: httpkit helpers
- docs: HTTP helpers usage

## v0.4.0 — 2026-02-03
- app: hooks (request start/end/panic) без внешних зависимостей
- tests: hooks integration contract
- docs: hooks usage

## v0.3.0 — 2026-02-03
- router: wildcard `*path` (catch-all) с фиксированной политикой
- routing docs: спецификация wildcard
- tests: интеграционные кейсы wildcard

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
