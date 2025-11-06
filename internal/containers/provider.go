package containers

import "errors"

var (
	ErrNotFound = errors.New("container not found")
)

// Container 描述容器基本資訊
// 在不同 Provider（Docker/K8s/Mock）之間以此為交換模型。
type Container struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Image     string `json:"image"`
	CreatedAt int64  `json:"createdAt"`
	Status    string `json:"status"` // created|running|stopped
}

// CreateOptions 建立容器所需參數。
type CreateOptions struct {
	Name  string `json:"name"`  // 可選
	Image string `json:"image"` // 必填
}

// Provider 提供容器操作抽象。
type Provider interface {
	Create(opts CreateOptions) (Container, error)
	Start(id string) error
	Stop(id string) error
	Delete(id string) error
}

// JobOptions 定義一次性作業的參數：將主機資料夾掛載進容器並執行命令。
type JobOptions struct {
    Image        string   `json:"image"`
    HostDir      string   `json:"hostDir"`
    ContainerDir string   `json:"containerDir"`
    Cmd          []string `json:"cmd"`
}

// JobRunner 可選介面：支援一次性作業。
type JobRunner interface {
    RunJob(opts JobOptions) (exitCode int64, logs string, err error)
}
