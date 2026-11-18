// file ini dicreate manual, bukan digenerate sqlc
package wallet

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/murfidaz/go-microservices-course/internal/wallet/walletdb"
)

type Posting struct {
	AccountID uuid.UUID
	Direction Direction
	Amount    int64
	Currency  string
}

type NewTransaction struct {
	Description string
	Postings    []Posting
}
type Transaction struct {
	walletdb.Transaction
	Entries []Entry `json:"entries"`
}

func (s *Store) inTx(ctx context.Context, fn func(q *walletdb.Queries) error) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin: %w", err)
	}

	defer tx.Rollback(ctx)

	if err := fn(s.q.WithTx(tx)); err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit: %w", err)
	}
	return nil
}

func (s *Store) PostTransaction(ctx context.Context, nt NewTransaction) (Transaction, error) {
	var posted Transaction
	err := s.inTx(ctx, func(q *walletdb.Queries) error {
		journal, err := q.CreateTransaction(ctx, nt.Description)
		if err != nil {
			return fmt.Errorf("insert transaction: %w", err)
		}
		posted.Transaction = journal

		for _, p := range nt.Postings {
			entry, err := q.CreateEntry(ctx, walletdb.CreateEntryParams{
				TransactionID: journal.ID,
				AccountID:     p.AccountID,
				Direction:     p.Direction,
				Amount:        p.Amount,
				Currency:      p.Currency,
			})
			if err != nil {
				return fmt.Errorf("insert entry: %w", err)
			}
			posted.Entries = append(posted.Entries, entry)
		}
		return nil
	})

	if err != nil {
		return Transaction{}, fmt.Errorf("post transaction: %w", err)
	}
	return posted, nil
}
