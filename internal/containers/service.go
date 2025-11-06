package containers

import (
	"os"

    "container-manager/internal/storage"
)

// Service 封裝 Provider 與資料持久化。
type Service struct {
	provider Provider
	repo     *storage.ContainerRepository
}

func NewService() *Service {
	p := os.Getenv("PROVIDER")
	var prov Provider
	switch p {
	case "docker":
		prov = NewDockerProvider()
	default:
		prov = NewMockProvider()
	}

	db, _ := storage.OpenDefault()
	_ = storage.Migrate(db)
	repo := storage.NewContainerRepository(db)

	return &Service{provider: prov, repo: repo}
}

// NewServiceWith 允許在測試中注入 provider 與 repository。
func NewServiceWith(provider Provider, repo *storage.ContainerRepository) *Service {
    return &Service{provider: provider, repo: repo}
}

func (s *Service) Create(opts CreateOptions) (Container, error) {
	c, err := s.provider.Create(opts)
	if err != nil {
		return Container{}, err
	}
	_ = s.repo.Create(storage.ContainerRecord{ID: c.ID, Name: c.Name, Image: c.Image, Status: c.Status, CreatedAt: c.CreatedAt})
	return c, nil
}

func (s *Service) Start(id string) error {
	if err := s.provider.Start(id); err != nil { return err }
	_ = s.repo.UpdateStatus(id, "running")
	return nil
}

func (s *Service) Stop(id string) error {
	if err := s.provider.Stop(id); err != nil { return err }
	_ = s.repo.UpdateStatus(id, "stopped")
	return nil
}

func (s *Service) Delete(id string) error {
	if err := s.provider.Delete(id); err != nil { return err }
	_ = s.repo.UpdateStatus(id, "deleted")
	return nil
}

// RunJob 如果底層 provider 支援 JobRunner，則執行一次性作業。
func (s *Service) RunJob(opts JobOptions) (int64, string, error) {
    if jr, ok := s.provider.(JobRunner); ok {
        return jr.RunJob(opts)
    }
    return 0, "", ErrNotFound // 使用通用錯誤；也可改為自定義錯誤
}
