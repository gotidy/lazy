# lazy

Starts something lazy.

```go
func ConnectDB(ctx context.Context) (*sql.DB, error) {
    db, err := sql.Open("postgres", "postgres://...")
    if err!= nil {
        return nil, err
    }
    return db, nil
}
createDB := lazy.Me(ctx, ConnectDB, WithRetry(iters.Trim(iters.Repeat(time.Second), 5))
...
v, err := connectDB(ctx)
```
