# go-microservices-course

membuat aplikasi microservices menggunakan golang Tanpa Framework dan ORM, karena dengan golang saja sudah powerfull.

apa saja yang didapat menggunakan pattern microservices?

| Keunggulan | Kekurangan | Solusinya |
| :--- | :--- | :--- |
| independent deploys and scaling | far more moving parts to run | Docker Compose now, Kubernetes |
| each service owns its data | no transacion can span two services | an outbox, sagas |
| failures stay contained | the network fails: timeouts, retries, duplicates | deadline, idempotency |
| clear boundaries | one request, logs in five places | distributed tracing |

## Jika memulai sebuah produk, mulailah dengan teknik Modular Monolith
pecah bagian yang perlu dipecah sajago mod , berikut batasan yang kita pecah pertamakali: identity, the ledger, money movement, notifications

### Pecahkan berdasarkan bisnis proses, bukan berdasarkan hal teknis
misalkan auth, wallet, transfers, notifications masing-masing memiliki API, Data dan Events. jadi tidak ada service yang bisa membaca table service lain

# Stack 
## Go 1.26
- net/http
- log/slog
- context

kenapa golang? karena golang itu:
- **A single static binary**: dibuild menjadi 1 binary yang mana tanpa runtime, tanpa virtual machine
- **Concurrency in the language**: support concurrency menggunakan goroutines yang tidak banyak memakan resource, serta context yang membawa deadline dan cancellation
- **A strandard library for servers**: golang support library untuk server seperti, net/http, encoding/json, crypto, log/slog
- **Boring, on purpose** golang membosankan karena sengaja, soalnya bahasa golang itu kecil, satu format, cepat dibuild, dan kompabilitas dari versi golang yang lama masih dijamin
- **The language of the cloud** golang merupakah bahasa yang disukai oleh aplikasi cloud, seperti Docker, Kubernetes, Prometheus, Jaeger dan NATS itu ditulis menggunakan Golang, jadi bahasa golang sangat menjamin

### Kekurangan Golang
**The trade-off: explicit errors**: karena golang tidak mempunyai exception (try catch) maka sering menggunakan **if err != nil** dibanyak kodingan, dan setiap kodingan yang akan memungkikan terjadi error maka wajib menggunakannya

### Alternative Golang
- **Java - Kotlin - C#**: bahasa yang sudah lama digunakan, akan tetapi menggunakan runtimes yang berat dan banyak sekal framworks
- **Node**: framework javascript yang awalnya cepat, akan tetapi CPU-heavy code
- **Rust**: lebih banyak kendali, lebih banyak jaminan akan tetapi proses pembelajaranya lebih sulit untuk dipahami developer junior

## gRPC
- between services
- protobuf

## PostgreSQL
- the ledger
- pgx
- sqlc

## Redis
- rate limits
- idempotency cache

## NATS JetStream
- durable streams of events

## Docker
untuk image dan local infra(develop)

## k6
untuk load test

## Kubernetes
untuk deploy production

# Tools Observability
## OpenTelemetry
trace from every hop

## Jaeger
untuk trace aplikasi

## Mailpit
aplikasi untuk fake email ketika development


## Prometheus
aplikasi untuk collect metric, scraped


## Grafana
aplikasi untuk menampilkan data metric


# Struktur File & Folder
```
root/
├── cmd/
│   └── wallet/
│       └── main.go
├── db/
├── deploy/
├── docker-compose.yaml
├── go.mod
├── Makefile
├── README.md
└── sqlc.yaml
```
- ```go.mod``` file config golang
- ```docker-compose.yaml``` digunakan untuk infra ketika develop aplikasi, dan juga sebagai dokumentasi infra apa saja yang kita gunakan
- ```main.go``` file utama setiap service
- ```db``` merupakan folder yang berhubungan dengan database postgres, seperti file migration
- ```deploy``` merupakan folder yang digunakan ketika docker compose dijalankan atau digunakan juga ketika ci/cd
- ```Makefile``` file script untuk memudahkan cmd
- ```sqlc.yaml``` file config untuk file sqlc


# Mulai

## Episode 1
### Graceful shutdown (penting!!!)
memberhentikan dengan aman adalah bagian dari pekerjaan setiap aplikasi
- **Deploy, scale-downs, drained nodes**: aplikasi pertamakali akan menanyakan dengan sinyal SIGTERM
  - **Stop accepting**: aplikasi berhenti menerima request dengan cara tutup listener: no new connection
  - **Finish what's in flight**: aplikasi akan menunggu setiap request selesai prosesnya
  - **Close the rest, and exit**: aplikasi akan menutup semua koneksi database pool, nats dll

## Episode 2
### Double-entry bookkeeping
adalah metode yang sudah ada sejak 500tahun yang lalu dengan cara **setiap pergerakan keuangan dicatat dua kali, ketika uang diterima dan ketika uang dikeluarkan**

| Transaksi | Account | Debit | Credit |
| :--- | :--- | :--- | :--- |
| Transaki 1 - eko deposit 50.000 | Bank - Asset | 50.000 | |
| | Dompet Eko - liablility | | 50.000|
| Transaki 2 - eko bayar diki 2.000 | Dompet Eko - liability | 2.000 | |
| | Dompet Diki - liablility | | 2.000|

**di setiap transaksi: total debits = total credits**

### Where does the balance live?
Sebuah angka yang dapat diubah atau saldo yang diperoleh dari entri-entri

#### A balance column (penggunaan yang salah)
```
UPDATE wallets
SET ballance = ballance - 2000
WHERE id = $1;
```
kenapa salah:
- concurrent update bisa saling overwrite
- tidak adanya riwayat: misalkan kenapa saldo sisa 42.00?
- sebuah bug tersembuyi 
- tidak bisa reconcile ulang

#### Derived from entries (penggunan yang benar)
```
SELECT sum(CASE direction WHEN 'credit' THEN amount ELSE -amount END)
FROM entries
WHERE account_id =  $1;
```
kenapa benar:
- Append-only: entries tidak pernah berubah
- Setiap rupiah dijelaskan berdasarkan entri
- Kesalahan diperbaiki dengan membalikkan transaksi (reversing transactions)
- Cache dan snapshop bisa selalu dibuild ulang

### Money is an integer
pada video menjelaskan untuk tipe data uang itu lebih baik menggunakan integer

- contoh jika menggunakan float64 ```0.10 + 0.20 = 0.30000000000000004``` nilainya tidak sempurna menjadi 0.3
- jadi simpanlah ke int64, contoh $19.99 disimpan ke int64 1999
- int64 mempunyai nilai max: 9,222,372,036,854,775,807
- postgres punya variabel numeric, akan tetapi golang tidak mempunyai variable decimal

## Episode 3
### Three ways to talk to Postgres from Go
- database/sql by hand (tidak dianjurkan): karena terlalu banyak kodingan dan kemungkinan typo lebih tinggi, serta terlalu banyak boilerplate
- an ORM (tidak dianjurkan): karena menyembunyikan query sql, N+1 query dan join, penggunaan vies, trigger dan CTEs jadi lebih sulit
- **sqlc** (sangat dianjurkan): karena kita yang membuat sql query, bisa mengecek schema ketika kita generate, murni menggunakan golang(no reflection, no magic)

### How sqlc works
sqlc membaca file .sql yang dibuat dari /migrations dan /queries dan kemudian digenerate menjadi bahasa golang

### Why a connection pool
sebuah koneksi baru ke postgres memakan biaya TC Handsake, auth dan sebuah backend proses, jadi untuk menghemat biaya kita menggunakan pool, berikut konfigurasi yang baik
- MinConns 2: minimal koneksi sebanyak 2
- MaxConns 10: maksimal koneksi sebanyak 10
- MaxConnLifetime 1h + 5m jitter: koneksi akan di recycle setelah 1 jam + random antar 5 menit
- MaxConIdelTime 15m: koneksi akan tertutup 15menit jika idle(tidak digunakan)
- HealthCheckPerios 30s: pengecekan koneksi setiap 30 detik
- ConnectionTimeout 5s: jika terjadi kendala, koneksi akan tertutup dalam 5detik

jadi dalam membuatkan koneksi harus diperhatikan juga
```
replicas x MaxConns must stay below max_connection (default postgres 100)
```
jadi misalkan service wallet mempunyai 5 replica maka akan dibutuhkan 5 x 10 = 50 koneksi

### Layer in the wallet service
```
cmd/wallet -> wallet.Store -> walletdb.Queries -> pgxpool.Pool -> PostgreSQL
```
- cmd/wallet = main: configuration, wiring, the HTTP server
- wallet.Store = out repository: domain types, errors, database transaction
- walletdb.Queries = generated by sqlc - one method per query
- pgxpool.Pool = connections - internal/platform/postgres
- PostgreSQL = the wallet database - constrains have the final word

### Posting a transaction atomically
```
BEGIN -> INSERT -> INSERT x N -> COMMIT
```

- **inTx(ctx, fn)** = begin - run fn with queries bound to the transaction - commit when fn return nill - roll back on an error or a panic
- **Never mix the pool and the transaction inside fn** = a query on the pool runs outside the transaction, on another connection, and can't see its uncommitted rows


## Episode 4
Config, Logging & Errors
- struktur konfigurasi sistem/env
- struktur log dengan request id
- middleware
- error dengan informasi yang jelas

### A middleware chain
setiap request melewati **Layer** yang sama sebelum sampai ke sebuah handler
```
# Layer
RequestID -> AccessLog -> Recover -> ServerMux -> handler
```
- RequestID = menerima atau membuat X-Request-ID, kemudian dimasukan kedalam context
- AccessLog = method, path, status, duration, bytes
- Recover = sebuah panic menjadi kode 500 dengan JSON error
- ServeMux -> handler = GET /healthz, GET /readyz, dan lain-lain seperti berupa JSON 404

**Middleware** itu adalah sebuah function, func(http.Handler) http.Handler, yang membungkus handler selanjutnya dan menjalankan semua kodingan sebelum dan setelahnya

**Order matters**
layer terluar melihat request pertama kali dan response dibagian akhir, access log harus bisa melihat error 500 yang meRecover writes

### Configuration from the environment
build yang sama berjalan dimanapun, hanya env yang berubah

### Structured logging with log/slog
key-value pairs daripada beberapa kalimat, jadi mesin bisa mencari mereka

- pada Development log berupa TEXT
- pada Production log berupa JSON
- Levels = DEBUG, INFO, WARN, ERROR, dipilih berdasarkan LOG_LEVEL
- Base Attributes = service dan versi pada setiap baris log
- Handlers do the writings = sebuah format logger akan dibuat dan dibungkus oleh handler

### Request IDs
satu ID per request, pada setipa baris log yang terjadi, didalam setiap service yang tersentuh
```
Client -> Gateway -> Transfer service/ Wallet service
```
- Client misalkan ketika POST /transfers
- Gateway akan membuat Request ID, misalkan X-Request-ID: 7Q2M...
- Transfer service & Wallet service akan menggunakan header dari gateway dengan id yang sama pada dalam log

jadi ketika kita mencari log untuk satu ID maka kita akan melihat seluruh riwayat yang terjadi, dan ketika client melaporkan error, request id pada body response akan membuat kita gampang mencarinya

### Errors become HTTP in exactly one place
Domain code mengatakan apa saja yang salah, dan HTTP layer menentukan bagaimana mengatakannya
| Error | Status | Code |
| :--- | :--- | :--- |
| apperr.ErrNotFound | 404 | not_found |
| apperr.ErrConflict | 409 | conflict |
| *apperr.ValidationError | 422 | invalid_input |
| malformed JSON body | 400 | bad_request |
| body over 1 MiB | 413 | body_too_large |
| anything else | 500 | internal |
