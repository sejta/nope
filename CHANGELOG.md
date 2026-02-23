# Changelog

Все заметные изменения фиксируются здесь.

## v1.6.3 — 2026-02-23
- server: добавлен `Validate()` для явной предзапускной проверки конфигурации
- server: `RunContext` теперь валидирует конфиг до запуска
- tests/docs: добавлены тесты и README-пример с `srv.Validate()`

## v1.6.2 — 2026-02-23
- server: в `safeHandle` сохраняется текст panic-причины роутера в `buildErr`
- tests: проверка, что ошибка регистрации содержит конкретную причину (`router: ...`)

## v1.6.1 — 2026-02-23
- app: `Run` возвращает `ErrNilHandler` вместо panic при `nil` handler
- httpkit: добавлен safe-адаптер `TryAdapt` (`error` вместо panic)
- server: регистрация невалидных route pattern переводится в `buildErr` вместо panic
- tests: добавлены кейсы на safe-поведение для `app/httpkit/server`

## v1.6.0 — 2026-02-23
- server: новый фасад для быстрого старта (`New`, `Group`, `Use`, `Run`)
- tests/docs: покрытие фасада и quick start в README
- examples: добавлен `examples/facade`

## v1.5.0 — 2026-02-07
- middleware: optional CORS
- tests/docs: CORS middleware

## v1.4.0 — 2026-02-07
- clientkit: JSON helpers + HTTPError + limits
- tests/docs: clientkit

## v1.3.0 — 2026-02-07
- dbkit: QueryAll/QueryOne/Exists helpers
- dbkit: ExecAffected/ExecOne/ExecAtMostOne helpers
- tests/docs: dbkit storage helpers

## v1.2.0 — 2026-02-06
- obs: hooks-based access log + metrics callbacks
- middleware: GetRequestID helper
- docs: OBSERVABILITY.md

## v1.1.0 — 2026-02-06
- docs: MIDDLEWARE.md (recommended order + contracts)
- httpkit: middleware.TimeoutError (optional responder)
- tests/docs: timeout middleware coverage + README snippet

## v1.0.0 — 2026-02-03
- First stable release. No functional changes since v0.9.0.

## v0.9.0 — 2026-02-03
- release candidate: API stable statement
- docs: final sweep + API.md stability
- ci: GitHub Actions uses scripts/ci.sh
- no functional changes

## v0.8.0 — 2026-02-03
- stabilization: public API inventory, godoc, docs sync
- examples: smoke-build via scripts/ci.sh
- no behavior changes

## v0.7.0 — 2026-02-03
- dbkit: MySQL-first Open / Conn / InTx
- dbkit: error classification (no rows / unique / fk)
- docs: DBKit usage

## v0.6.0 — 2026-02-03
- policy freeze: routing/app
- recorder: Unwrap для совместимости обёрток
- panic-after-write: поведение закреплено

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
