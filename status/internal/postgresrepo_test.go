package statuses

import (
	"fmt"
	"github.com/DATA-DOG/go-txdb"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
	"time"
	"yatc/internal"
	statuses "yatc/status/pkg"
)

func init() {
	databaseConn := os.Getenv("DATABASE_CONNECTION_STRING")
	if databaseConn != "" {
		txdb.Register("psql_txdb", "postgres", databaseConn)
	}
}

func prepare(t *testing.T) (repo Repository, cleanup func() error) {
	if os.Getenv("DATABASE_CONNECTION_STRING") == "" {
		t.Skipf("set DATABASE_CONNECTION_STRING to run this integration test")
	}

	cName := fmt.Sprintf("connection_%d", time.Now().UnixNano())
	db, err := sqlx.Open("psql_txdb", cName)

	if err != nil {
		t.Fatalf("open psql_txdb connection: %s", err)
	}

	schema := `CREATE TABLE IF NOT EXISTS statuses (
			id UUID PRIMARY KEY,
			content TEXT,
			user_id UUID
		);`

	_, err = db.Exec(schema)
	if err != nil {
		t.Fatalf("cant apply default scheme to database %s", err)
	}

	return NewPostgresRepo(db), db.Close
}

func TestPostgresRepo_Create(t *testing.T) {
	t.Parallel()
	postgresRepo, cleanup := prepare(t)
	defer func() {
		_ = cleanup
	}()
	createStatus := statuses.Status{Id: uuid.New(), Content: "test status", UserId: uuid.New()}

	createdStatus, err := postgresRepo.Create(createStatus)
	if err != nil {
		assert.NoError(t, err)
	}
	assert.Equal(t, createStatus, createdStatus)
}

func TestPostgresRepo_Get(t *testing.T) {
	t.Parallel()
	postgresRepo, cleanup := prepare(t)
	defer func() {
		_ = cleanup
	}()
	createStatus := statuses.Status{Id: uuid.New(), Content: "test status", UserId: uuid.New()}

	createdStatus, err := postgresRepo.Create(createStatus)
	if err != nil {
		assert.NoError(t, err)
	}
	assert.Equal(t, createStatus, createdStatus)

	gotStatus, err := postgresRepo.Get(createStatus.Id)
	if err != nil {
		assert.NoError(t, err)
	}
	assert.Equal(t, createStatus, gotStatus)
}

func TestPostgresRepo_List(t *testing.T) {
	t.Parallel()
	postgresRepo, cleanup := prepare(t)
	defer func() {
		_ = cleanup
	}()
	createStatus := statuses.Status{Id: uuid.New(), Content: "test status", UserId: uuid.New()}
	createStatus2 := statuses.Status{Id: uuid.New(), Content: "test status", UserId: uuid.New()}

	createdStatus, err := postgresRepo.Create(createStatus)
	if err != nil {
		assert.NoError(t, err)
	}
	assert.Equal(t, createStatus, createdStatus)

	createdStatus2, err := postgresRepo.Create(createStatus2)
	if err != nil {
		assert.NoError(t, err)
	}
	assert.Equal(t, createStatus2, createdStatus2)

	allStatus, err := postgresRepo.List()
	if err != nil {
		assert.NoError(t, err)
	}
	assert.Equal(t, 2, len(allStatus))
}

func TestPostgresRepo_Delete(t *testing.T) {
	t.Parallel()
	postgresRepo, cleanup := prepare(t)
	defer func() {
		_ = cleanup
	}()
	createStatus := statuses.Status{Id: uuid.New(), Content: "test status", UserId: uuid.New()}

	createdStatus, err := postgresRepo.Create(createStatus)
	if err != nil {
		assert.NoError(t, err)
	}
	assert.Equal(t, createStatus, createdStatus)

	deletedStatus, err := postgresRepo.Delete(createStatus.Id)
	if err != nil {
		assert.NoError(t, err)
	}
	assert.Equal(t, createStatus, deletedStatus)

	_, err = postgresRepo.Get(createStatus.Id)
	if err != nil {
		assert.ErrorIs(t, err, internal.NotFoundError(createStatus.Id))
	} else {
		assert.Failf(t, "expected error NotFound", "got error: %s", err)
	}
}
