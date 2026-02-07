# ClientKit

ClientKit — тонкий фасад над `net/http` для исходящих запросов.
Цели: меньше рутины, безопасные дефолты, без магии и без зависимостей.

## Пример: GET JSON

```go
client := clientkit.DefaultClient()

var out struct {
	ID    int    `json:"id"`
	Email string `json:"email"`
}

meta, err := clientkit.GetJSON(ctx, client,
	"https://api.example.com/users/1",
	&out,
	&clientkit.JSONOptions{DisallowUnknownFields: true},
)
if err != nil {
	return err
}
_ = meta
```

## Пример: POST JSON

```go
client := clientkit.DefaultClient()

req := struct {
	Email string `json:"email"`
}{Email: "x@example.com"}

var out struct {
	ID int `json:"id"`
}

_, err := clientkit.PostJSON(ctx, client,
	"https://api.example.com/users",
	&req,
	&out,
	&clientkit.JSONOptions{MaxBody: 1 << 20},
)
if err != nil {
	return err
}
```

## Пример: обработка HTTPError

```go
_, err := clientkit.GetJSON(ctx, client,
	"https://api.example.com/users/1",
	&out,
	nil,
)
if err != nil {
	if he, ok := clientkit.IsHTTPError(err); ok {
		// he.Status, he.Method, he.URL, he.Body
		return err
	}
	return err
}
```
