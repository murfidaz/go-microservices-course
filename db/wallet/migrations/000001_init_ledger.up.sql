BEGIN;

-- debit atau credit: tipe setiap entri
CREATE TYPE entry_direction AS ENUM('debit','credit');

-- account adalah semua yang bersangkutan dengan keuangan:
-- wallet pengguna, atau sistem account seperti account bank yang dimana deposit berasal
CREATE TABLE accounts(
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    -- user dari service auth, NULL jika sistem accounts
    owner_id uuid,
    name text NOT NULL CHECK (name <> ''),
    currency text NOT NULL CHECK (currency ~ '^[A-Z]{3}$'),
    normal_balance entry_direction NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE (id, currency)
);

CREATE INDEX accounts_owner_id_idx ON accounts (owner_id);

-- journal: setiap baris per pergerakan keuangan(money)
CREATE TABLE transactions(
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    description text NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now()
);

-- setiap baris sebuah transaksi: sebuah amount yang didebitkan atau kreditkan ke sebuah account
CREATE TABLE entries(
    id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    transaction_id uuid NOT NULL REFERENCES transactions (id),
    account_id uuid NOT NULL,
    direction entry_direction NOT NULL,
    amount bigint NOT NULL CHECK (amount > 0),
    currency text NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    -- sebuah entries currency harus sama dengan currency accountnya
    FOREIGN KEY (account_id, currency) REFERENCES accounts(id, currency)
);

-- postgres secara bawaan tidak men-index FOREIGN KEY
CREATE INDEX entries_transaction_id_idx ON entries (transaction_id);
CREATE INDEX entries_account_id_idx ON entries (account_id);

-- sebuah transaction haru memiliki entri, dan data debitnya harus sama dengan data creditnya
-- di setiap currency. Cek ketika COMMIT, ketika semua entries ada.
-- maka dibuatkanlah trigger
CREATE FUNCTION check_transaction_balanced() RETURNS trigger
LANGUAGE plpgsql AS $$
DECLARE
    tx_id uuid;
BEGIN
    IF TG_TABLE_NAME = 'transactions' THEN
        tx_id := NEW.id;
    ELSE
        tx_id := NEW.transaction_id;
    END IF;

    IF NOT EXISTS (SELECT 1 FROM entries WHERE transaction_id = tx_id) THEN
        RAISE EXCEPTION 'transaction % has no entries', tx_id
            USING ERRCODE = 'check_validation';
    END IF;

    IF EXISTS (
        SELECT 1
        FROM entries
        WHERE transaction_id = tx_id
        GROUP BY currency
        HAVING sum(CASE direction WHEN 'debit' THEN amount ELSE -amount END) <> 0
    ) THEN
        RAISE EXCEPTION 'transaction % is not balanced', tx_id
            USING ERRCODE = 'check_validation';
    END IF;

    RETURN NULL;
END;
$$;

-- jalankan trigger ketika insert ke table transactions
CREATE CONSTRAINT TRIGGER transaction_balanced
    AFTER INSERT ON transactions
    DEFERRABLE INITIALLY DEFERRED
    FOR EACH ROW EXECUTE FUNCTION check_transaction_balanced();

-- jalankan trigger ketika insert ke table entries
CREATE CONSTRAINT TRIGGER entries_balanced
    AFTER INSERT ON entries
    DEFERRABLE INITIALLY DEFERRED
    FOR EACH ROW EXECUTE FUNCTION check_transaction_balanced();

-- data ledger hanya ditambah (append): jika terjadi kesalahan maka perbaikannya dengan mereverse transaction
CREATE FUNCTION forbid_ledger_changes() RETURNS trigger
LANGUAGE plpgsql AS $$
BEGIN
    RAISE EXCEPTION '% is append-only, post a reversing transaction instead', TG_TABLE_NAME
        USING ERRCODE = 'restrict_violation';
END;
$$;

-- jalankan trigger ketika UPDATE atau DELETE transactions
CREATE TRIGGER transactions_append_only
    BEFORE UPDATE OR DELETE ON transactions
    FOR EACH ROW EXECUTE FUNCTION forbid_ledger_changes();

-- jalankan trigger ketika UPDATE atau DELETE entries
CREATE TRIGGER transactions_append_only
    BEFORE UPDATE OR DELETE ON entries
    FOR EACH ROW EXECUTE FUNCTION forbid_ledger_changes();

-- Balances itu dibuat dari entries dan tidak pernah disimpan, jadi gunakan view.
CREATE VIEW account_balances AS
SELECT
    a.id AS account_id,
    a.name,
    a.currency,
    coalesce(sum(
        CASE WHEN e.direction = a.normal_balance THEN e.amount ELSE -e.amount END
    ), 0)::bigint AS balance
FROM accounts a
LEFT JOIN entries e ON e.account_id = a.id
GROUP BY a.id;

COMMIT;
