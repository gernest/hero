package hero

import (
	"fmt"
	"os"
	"testing"

	"github.com/jinzhu/gorm"
	_ "github.com/lib/pq"
)

//dbConn us the database connection used for testing purposes. If the connection is initialised isOpen will be set
// to true,
var dbConn = struct {
	isOpne bool
	db     *gorm.DB
}{}

// this is the main server instance. It should be initialized and available to all tests in this package.
var testServer *Server

func TestMain(m *testing.M) {
	if env := os.Getenv("DB_CONN"); env != "" {
		dialect := os.Getenv("DB_DIALECT")
		db, err := gorm.Open(dialect, env)
		if err != nil {
			fmt.Printf("hero: some tests wont run due to bad database connection %v \n", err)
		} else {
			dbConn.isOpne = true
			dbConn.db = &db
			config := DefaultConfig()

			testServer = NewServer(config, &SimpleTokenGen{}, nil)
			testServer.Migrate()
		}
	}
	status := m.Run()
	if dbConn.isOpne {
		testServer.DropAllTables()
		dbConn.db.Close()
		testServer.q.Close()
	}
	os.Exit(status)
}
