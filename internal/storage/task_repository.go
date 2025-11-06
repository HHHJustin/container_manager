package storage

import (
    "database/sql"
    "encoding/json"
    "time"
)

type TaskStatus string

const (
    TaskPending   TaskStatus = "pending"
    TaskRunning   TaskStatus = "running"
    TaskSucceeded TaskStatus = "succeeded"
    TaskFailed    TaskStatus = "failed"
)

type ContainerTask struct {
    ID          string
    ContainerID string
    CmdJSON     string
    Status      TaskStatus
    ExitCode    int
    Logs        string
    CreatedAt   time.Time
    FinishedAt  sql.NullTime
}

type TaskRepository struct{ db *sql.DB }

func NewTaskRepository(db *sql.DB) *TaskRepository { return &TaskRepository{db: db} }

func (r *TaskRepository) Insert(containerID string, cmd []string) (string, error) {
    b, _ := json.Marshal(cmd)
    id := time.Now().UTC().Format("20060102T150405.000Z0700")
    _, err := r.db.Exec(`INSERT INTO container_tasks(id, container_id, cmd_json, status, created_at) VALUES($1,$2,$3,$4,$5)`, id, containerID, string(b), string(TaskPending), time.Now().Unix())
    return id, err
}

func (r *TaskRepository) UpdateResult(id string, status TaskStatus, exitCode int, logs string) error {
    _, err := r.db.Exec(`UPDATE container_tasks SET status=$1, exit_code=$2, logs=$3, finished_at=$4 WHERE id=$5`, string(status), exitCode, logs, time.Now().Unix(), id)
    return err
}


