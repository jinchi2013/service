package tests

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"fmt"
	"io"
	"os"
	"testing"
	"time"

	"github.com/jinchi2013/service/busniess/data/schema"
	"github.com/jinchi2013/service/busniess/data/store/user"
	"github.com/jinchi2013/service/busniess/sys/auth"
	"github.com/jinchi2013/service/busniess/sys/database"
	"github.com/jinchi2013/service/foundation/docker"
	keystore "github.com/jinchi2013/service/foundation/keyStore"
	"github.com/jinchi2013/service/foundation/logger"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

// Success and failure markers.
const (
	Success = "\u2713"
	Failed  = "\u2717"
)

type DBContainer struct {
	Image string
	Port  string
	Args  []string
}

// NewUnit creates a test database inside a Docker container. It creates the
// required table structure but the database is otherwise empty. It returns
// the database to use as well as a function to call at the end of the test.
func NewUnit(t *testing.T, dbc DBContainer) (*zap.SugaredLogger, *sqlx.DB, func()) {
	r, w, _ := os.Pipe() // Pipe gives us a reader and writer
	old := os.Stdout
	os.Stdout = w // Now everything in Stdout would be saved into the writer

	// start the docker container
	c := docker.StartContainer(t, dbc.Image, dbc.Port, dbc.Args...)

	// connecting to database
	db, err := database.Open(database.Config{
		User:       "postgres",
		Password:   "postgres",
		Name:       "postgres",
		Host:       c.Host,
		DisableTLS: true,
	})

	if err != nil {
		t.Fatalf("Opening database connection: %v", err)
	}

	t.Log("Waiting for database to be ready ... ")

	// create a timer for 10s to wait the migrate and seed the data
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// start to migrate the db
	if err := schema.Migrate(ctx, db); err != nil {
		// if there is any error
		// drop all the logs, and stop the container
		docker.DumpContainerLogs(t, c.ID)
		docker.StopContainer(t, c.ID)
		t.Fatalf("Migrating error: %s", err)
	}

	// seeding the data into database
	if err := schema.Seed(ctx, db); err != nil {
		// if there is any error
		// drop all the logs, and stop the container
		docker.DumpContainerLogs(t, c.ID)
		docker.StopContainer(t, c.ID)
		t.Fatalf("Seeding error: %s", err)
	}

	log, err := logger.New("TEST")
	if err != nil {
		t.Fatalf("logger error: %s", err)
	}

	// tear down is the function that should be invoked when the caller is done
	// with the database
	teardown := func() {
		t.Helper()
		db.Close()
		docker.StopContainer(t, c.ID)

		log.Sync()

		w.Close()
		var buf bytes.Buffer
		io.Copy(&buf, r) // reader, read from what writer writes, and save in the buffer
		os.Stdout = old  // give stdout its original writer
		fmt.Println("*************************** LOGS ***************************")
		fmt.Print(buf.String())
		fmt.Println("*************************** LOGS ***************************")
	}

	return log, db, teardown
}

// Test owns state for running and shutting down tests.
type Test struct {
	DB       *sqlx.DB
	Log      *zap.SugaredLogger
	Auth     *auth.Auth
	Teardown func()

	t *testing.T
}

// NewIntegration creates a database, seeds it, constructs an authenticator
func NewIntegration(t *testing.T, dbc DBContainer) *Test {
	log, db, teardown := NewUnit(t, dbc)

	// Create RSA keys to enable authentication in our service
	keyID := "4754d86b-7a6d-4df5-9c65-224741361492"
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}

	auth, err := auth.New(keyID, keystore.NewMap(map[string]*rsa.PrivateKey{keyID: privateKey}))
	if err != nil {
		t.Fatal(err)
	}

	test := Test{
		DB:       db,
		Log:      log,
		Auth:     auth,
		t:        t,
		Teardown: teardown,
	}

	return &test
}

func (test *Test) Token(email, pass string) string {
	test.t.Log("Generating token for test ...")

	store := user.NewStore(test.Log, test.DB)
	claims, err := store.Authenticate(context.Background(), time.Now(), email, pass)
	if err != nil {
		test.t.Fatal(err)
	}

	token, err := test.Auth.GenerateToken(claims)
	if err != nil {
		test.t.Fatal(err)
	}

	return token
}

func StringPointer(s string) *string {
	return &s
}

func IntPointer(i int) *int {
	return &i
}
