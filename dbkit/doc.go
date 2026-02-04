// Package dbkit предоставляет тонкую обёртку над database/sql.
//
// Политика:
// - MySQL — default и гарантированная классификация ошибок.
// - PostgreSQL — best‑effort (через SQLState()).
// - Другие драйверы → KindUnknown.
package dbkit
