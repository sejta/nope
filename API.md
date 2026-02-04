# Public API (v0.x)

Этот документ описывает **публичные** пакеты и стабильность API.

Политики роутинга и app зафиксированы с `v0.6.0`:
- `ROUTING.md` — routing policy (frozen)
- `DESIGN.md` — app policy (frozen)

Цель: стабилизировать API и заморозить в `v1.0`.

---

## Stability

Начиная с `v0.9.0` публичный API считается стабильным.

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

### json
Низкоуровневые JSON‑helpers (strict decode/encode).

### dbkit
Тонкая работа с `database/sql`: `Open`, `Conn`, `InTx`, классификация ошибок.

### spa
Static + fallback (опционально).

---

## Что не является public API

- `internal/*`
- тестовые пакеты
- неэкспортируемые символы внутри публичных пакетов
