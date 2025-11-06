package storage

import (
	"database/sql"
)

type ContainerRecord struct {
	ID        string
	Name      string
	Image     string
	Status    string
	CreatedAt int64
}

type ContainerRepository struct{ db *sql.DB }

func NewContainerRepository(db *sql.DB) *ContainerRepository { return &ContainerRepository{db: db} }

func (r *ContainerRepository) Create(rec ContainerRecord) error {
    _, err := r.db.Exec(`INSERT INTO containers(id,name,image,status,created_at) VALUES($1,$2,$3,$4,$5)`, rec.ID, rec.Name, rec.Image, rec.Status, rec.CreatedAt)
	return err
}

func (r *ContainerRepository) UpdateStatus(id, status string) error {
    _, err := r.db.Exec(`UPDATE containers SET status=$1 WHERE id=$2`, status, id)
	return err
}
