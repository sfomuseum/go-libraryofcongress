package libraryofcongress

import (
	"context"
	"database/sql"
	"fmt"
	"io/ioutil"
	"os"
	"sync"

	_ "modernc.org/sqlite"
)

// type Catalog is a struct used deduplicate IDs seen in the various LoC authority files.
// It is necessary specifically for the LCNAF file which is so big that tracking IDs in memory
// trigger "out of memory" errors so instead we track "attendance" on disk using a temporary SQLite
// database.
type Catalog struct {
	// path is the path to the temporary SQLite database on disk.
	path string
	// db is the `sql.DB` instance mapped to the temporary SQLite database on disk.
	db *sql.DB
	// mu is an internal `sync.RWMutex` instance used to prevent race conditions.
	mu *sync.RWMutex
}

// NewCatalog() returns a new `Catalog` instance configured by 'uri' which is expected to take
// the form of:
//
//	tmp://
func NewCatalog(ctx context.Context, uri string) (*Catalog, error) {

	tmpfile, err := ioutil.TempFile("", "catalog")

	if err != nil {
		return nil, fmt.Errorf("Failed to create temp file, %w", err)
	}

	tmpfile.Close()

	path := tmpfile.Name()

	dsn := fmt.Sprintf("%s", path)

	db, err := sql.Open("sqlite", dsn)

	if err != nil {
		return nil, fmt.Errorf("Failed to open database, %w", err)
	}

	pragma := []string{
		"PRAGMA JOURNAL_MODE=OFF",
		"PRAGMA SYNCHRONOUS=OFF",
		"PRAGMA LOCKING_MODE=EXCLUSIVE",
		"PRAGMA PAGE_SIZE=4096",
		"PRAGMA CACHE_SIZE=1000000",
	}

	for _, p := range pragma {

		_, err = db.Exec(p)

		if err != nil {
			return nil, fmt.Errorf("Failed to set %s, %w", p, err)
		}
	}

	_, err = db.ExecContext(ctx, `CREATE TABLE seen(id TEXT PRIMARY KEY);`)

	if err != nil {
		return nil, fmt.Errorf("Failed to create database table, %v", err)
	}

	mu := new(sync.RWMutex)

	c := &Catalog{
		path: path,
		db:   db,
		mu:   mu,
	}

	return c, nil
}

// ExistsOrStore() adds 'id' to the underlying SQLite database if it does not already exist.
func (c *Catalog) ExistsOrStore(ctx context.Context, id string) (bool, error) {

	c.mu.Lock()
	defer c.mu.Unlock()

	exists, err := c.Exists(ctx, id)

	if err != nil {
		return false, fmt.Errorf("Failed to determine whether %s exists, %w", id, err)
	}

	if exists {
		return true, nil
	}

	err = c.Store(ctx, id)

	if err != nil {
		return false, fmt.Errorf("Failed to store %s, %w", id, err)
	}

	return false, nil
}

// Exists() returns a boolean value indicating whether or not 'id' exists in the temporary SQLite database.
func (c *Catalog) Exists(ctx context.Context, id string) (bool, error) {

	var count int

	row := c.db.QueryRowContext(ctx, "SELECT COUNT(id) FROM seen WHERE id = ?", id)
	err := row.Scan(&count)

	if err != nil && err != sql.ErrNoRows {
		return false, fmt.Errorf("Failed to query for %s, %w", id, err)
	}

	if count == 0 {
		return false, nil
	}

	return true, nil
}

// Store() creates a new entry for 'id' in the temporary SQLite database.
func (c *Catalog) Store(ctx context.Context, id string) error {

	_, err := c.db.ExecContext(ctx, "INSERT INTO seen(id) VALUES(?)", id)

	if err != nil {
		return fmt.Errorf("Failed to insert for %s, %w", id, err)
	}

	return nil
}

// Close() removes the temporary SQLite database from disk.
func (c *Catalog) Close(ctx context.Context) error {
	return os.Remove(c.path)
}
