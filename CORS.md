# CORS middleware

CORS middleware нужен для браузерных клиентов и dev-сценариев.
По умолчанию ничего не разрешает — всё только через конфиг.

## Пример: простой allowlist

```go
cors := middleware.CORS(middleware.CORSOptions{
	AllowedOrigins: []string{"https://app.example"},
})

r := router.New()
// ... маршруты

h := cors(r)
_ = h
```

## Пример: credentials + expose headers

```go
cors := middleware.CORS(middleware.CORSOptions{
	AllowedOrigins:   []string{"https://app.example"},
	AllowCredentials: true,
	ExposedHeaders:   []string{"X-Trace"},
})
```

## Пример: preflight параметры

```go
cors := middleware.CORS(middleware.CORSOptions{
	AllowedOrigins: []string{"https://app.example"},
	AllowedMethods: []string{"GET", "POST"},
	AllowedHeaders: []string{"Content-Type", "Authorization"},
	MaxAge:         10 * time.Second,
})
```
