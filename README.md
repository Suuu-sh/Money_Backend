# Money_Backend

## Local Development

Run the API locally on a custom port (e.g. 8000):

```
PORT=8000 ./money-tracker
```

Or build first, then run:

```
go build -o money-tracker .
PORT=8000 ./money-tracker
```

The frontend can point to this backend by setting:

```
NEXT_PUBLIC_API_BASE_URL=http://localhost:8000/api
```

Notes:
- CORS allows `http://localhost:3000` by default (see `main.go`).
- The app uses SQLite (`moneytracker.db`) for local development.
