package sqlc

import (
	"database/sql"
	"log"
	"os"
	"testing"

	_ "github.com/lib/pq"
)

// the _ it use for package that we import to avoid when saving it gone
const (
	dbDriver = "postgres"
	dbSource = "postgresql://postgres:12345@localhost:5431/simple_bank?sslmode=disable"
)

var testQueries *Queries
var testDB *sql.DB

// this is to connect the test with db
func TestMain(m *testing.M) {
	var err error
	testDB, err = sql.Open(dbDriver, dbSource)
	if err != nil {
		log.Fatal("cannot connect to db")
	}
	testQueries = New(testDB)
	os.Exit(m.Run())

}
