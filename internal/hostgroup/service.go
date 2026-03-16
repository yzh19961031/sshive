package hostgroup

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
	Name string `json:"name" binding:"required,max=100"`
}

type UpdateReq struct {
	Name string `json:"name" binding:"required,max=100"`
}

func (s *Service) List(tenantID int64) ([]model.HostGroup, error) {
	return s.repo.List(tenantID)
}

func (s *Service) Create(tenantID int64, req CreateReq) (*model.HostGroup, error) {
	g := &model.HostGroup{TenantID: tenantID, Name: req.Name}
	if err := s.repo.Create(g); err != nil {
		return nil, fmt.Errorf("create hostgroup: %w", err)
	}
	return g, nil
}

func (s *Service) Update(tenantID, id int64, req UpdateReq) (*model.HostGroup, error) {
	g, err := s.repo.GetByID(tenantID, id)
	if err != nil {
		return nil, fmt.Errorf("group not found")
	}
	g.Name = req.Name
	if err := s.repo.Update(g); err != nil {
		return nil, fmt.Errorf("update hostgroup: %w", err)
	}
	return g, nil
}

func (s *Service) Delete(tenantID, id int64) error {
	return s.repo.Delete(tenantID, id)
}
