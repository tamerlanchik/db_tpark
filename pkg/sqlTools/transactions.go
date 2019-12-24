package sqlTools

import "database/sql"

// WithTransaction : wrapper who controls the transaction state
func WithTransaction(db *sql.DB, f func() error) error{
	tx, err := db.Begin()
	if err != nil {
		tx.Rollback()
		return err
	}
	err = f()
	if err != nil {
		tx.Rollback()
	} else {
		tx.Commit()
	}
	return err
}
