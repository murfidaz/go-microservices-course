
# Go Migrate
install golang migrate
```
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# cek migrate
which migrate
```

create migrate ke folder db berdasarkan database
```
migrate create -ext sql -dir db/wallet/migrations -seq init_ledger
```

menjalankan migrasi (development)
```
migrate -path db/wallet/migrations -database "postgres://payments:payments@localhost:5432/wallet?sslmode=disable" up
```

# Database
cek tabel pada database
```
docker compose exec -T postgres psql -U payments -d wallet -c '\dt'
```

cek tabel migrasi
```
docker compose exec -T postgres psql -U payments -d wallet -c 'TABLE schema_migrations'
```

# sqlc
install sqlc
```
go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
```

cek versi
```
sqlc version
```

generate sqlc, setelah meruba atau menambahkan file .sql
```
sqlc generate
```

# test
jalankan test
```
go test ./...
```

menggunakan make karena sudah dimasukan
```
make test
```

verbose dengan make
```
make test TESTFLAGS='-v -run Unbalance'
```

cek data 
```
docker compose exec -T postgres psql -U payments -d wallet -c 'SELECT direction, sum(amount) FROM entries GROUP BY 1 ORDER BY 1'
```
