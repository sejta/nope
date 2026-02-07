# Public API (v1.x)

Этот документ описывает **публичные** пакеты и стабильность API.

Политики роутинга и app зафиксированы с `v0.6.0`:
- `ROUTING.md` — routing policy (frozen)
- `DESIGN.md` — app policy (frozen)

Цель: поддерживать стабильность API в `v1.x`.

---

## Stability

Начиная с `v1.0.0` публичный API считается стабильным.

Breaking change — это:
- изменение сигнатур публичных функций/типов;
- изменение поведения, закреплённого в `ROUTING.md` и `DESIGN.md`.

Breaking‑изменения, если когда‑то понадобятся, будут выходить только в major‑версии (v2.0+).

---

## Public packages

### app
Bootstrap/runtime: запуск HTTP‑сервера, graceful shutdown, health, pprof, hooks.

### router
Минимальный роутер: static, `:param`, `*path`, `Mount`, 404/405 + Allow.

### errors
Единый error contract и JSON‑рендер ошибок.

### httpkit
Контракт handler’ов, адаптер, middleware и HTTP‑helpers.

### clientkit
Тонкий фасад для исходящих HTTP‑запросов: JSON helpers, HTTPError, лимиты.

### json
Низкоуровневые JSON‑helpers (strict decode/encode).

### dbkit
Тонкая работа с `database/sql`: `Open`, `Conn`, `InTx`, классификация ошибок, helpers `QueryAll/QueryOne/Exists` и `ExecAffected/ExecOne/ExecAtMostOne`.

### spa
Static + fallback (опционально).

### obs
Hooks-first observability: логирование и метрики через `app.Hooks`.

---

## Что не является public API

- `internal/*`
- тестовые пакеты
- неэкспортируемые символы внутри публичных пакетов
