package tests

import (
    "regexp"
    "testing"

    "container-manager/internal/containers"
    "container-manager/internal/storage"
    sqlmock "github.com/DATA-DOG/go-sqlmock"
)

func newRepoWithMock(t *testing.T) (*storage.ContainerRepository, sqlmock.Sqlmock) {
    t.Helper()
    db, mock, err := sqlmock.New()
    if err != nil {
        t.Fatalf("sqlmock: %v", err)
    }
    return storage.NewContainerRepository(db), mock
}

func TestService_Create_RecordsToRepo(t *testing.T) {
    repo, mock := newRepoWithMock(t)
    s := containers.NewServiceWith(containers.NewMockProvider(), repo)

    mock.ExpectExec(regexp.QuoteMeta("INSERT INTO containers(id,name,image,status,created_at) VALUES($1,$2,$3,$4,$5)")).WithArgs(sqlmock.AnyArg(), "demo", "alpine:3.20", "created", sqlmock.AnyArg()).WillReturnResult(sqlmock.NewResult(1, 1))

    c, err := s.Create(containers.CreateOptions{Name: "demo", Image: "alpine:3.20"})
    if err != nil { t.Fatalf("create: %v", err) }
    if c.Image != "alpine:3.20" { t.Fatalf("unexpected image: %s", c.Image) }

    if err := mock.ExpectationsWereMet(); err != nil { t.Fatalf("unmet expectations: %v", err) }
}

func TestService_Start_Stop_Delete_UpdateRepo(t *testing.T) {
    repo, mock := newRepoWithMock(t)
    s := containers.NewServiceWith(containers.NewMockProvider(), repo)

    mock.ExpectExec(regexp.QuoteMeta("INSERT INTO containers(id,name,image,status,created_at) VALUES($1,$2,$3,$4,$5)")).WithArgs(sqlmock.AnyArg(), "demo", "alpine:3.20", "created", sqlmock.AnyArg()).WillReturnResult(sqlmock.NewResult(1, 1))
    c, err := s.Create(containers.CreateOptions{Name: "demo", Image: "alpine:3.20"})
    if err != nil { t.Fatalf("create: %v", err) }

    mock.ExpectExec(regexp.QuoteMeta("UPDATE containers SET status=$1 WHERE id=$2")).WithArgs("running", c.ID).WillReturnResult(sqlmock.NewResult(1, 1))
    if err := s.Start(c.ID); err != nil { t.Fatalf("start: %v", err) }

    mock.ExpectExec(regexp.QuoteMeta("UPDATE containers SET status=$1 WHERE id=$2")).WithArgs("stopped", c.ID).WillReturnResult(sqlmock.NewResult(1, 1))
    if err := s.Stop(c.ID); err != nil { t.Fatalf("stop: %v", err) }

    mock.ExpectExec(regexp.QuoteMeta("UPDATE containers SET status=$1 WHERE id=$2")).WithArgs("deleted", c.ID).WillReturnResult(sqlmock.NewResult(1, 1))
    if err := s.Delete(c.ID); err != nil { t.Fatalf("delete: %v", err) }

    if err := mock.ExpectationsWereMet(); err != nil { t.Fatalf("unmet expectations: %v", err) }
}


