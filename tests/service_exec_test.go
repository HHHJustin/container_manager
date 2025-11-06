package tests

import (
    "testing"

    "container-manager/internal/containers"
)

type execMock struct{ containers.Provider }
func (e execMock) Exec(id string, cmd []string) (int, string, error) { return 0, "ok", nil }

func TestService_Exec(t *testing.T) {
    s := containers.NewServiceWith(execMock{}, nil)
    code, logs, err := s.Exec("cid", []string{"echo","hi"})
    if err != nil || code != 0 || logs != "ok" { t.Fatalf("unexpected: code=%d logs=%s err=%v", code, logs, err) }
}


