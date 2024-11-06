package sqlc

import (
	"context"
	"log"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/lib/pq"
)

// the _ it use for package that we import to avoid when saving it gone
const (
	dbSource = "postgresql://postgres:12345@localhost:5431/simple_bank?sslmode=disable"
)

var testStore Store

// this is to connect the test with db
func TestMain(m *testing.M) {
	var err error
	connPool, err := pgxpool.New(context.Background(), dbSource)
	if err != nil {
		log.Fatal("cannot connect to db")
	}
	testStore = NewStore(connPool)
	os.Exit(m.Run())

}
