package tenant

import (
	"fmt"

	"github.com/sshive/sshive/internal/model"
)

type Service struct {
	repo *Repo
}

func NewService() *Service {
	return &Service{repo: &Repo{}}
}

type CreateReq struct {
	Name string `json:"name" binding:"required,min=2,max=100"`
}

func (s *Service) Create(req CreateReq) (*model.Tenant, error) {
	t := &model.Tenant{Name: req.Name, Status: 1}
	if err := s.repo.Create(t); err != nil {
		return nil, fmt.Errorf("create tenant: %w", err)
	}
	return t, nil
}

func (s *Service) List(page, pageSize int) ([]model.Tenant, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	return s.repo.List(page, pageSize)
}
