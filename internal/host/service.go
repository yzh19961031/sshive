package host

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
	Name         string     `json:"name" binding:"required,max=100"`
	IP           string     `json:"ip" binding:"required"`
	Port         int        `json:"port"`
	AuthType     string     `json:"auth_type" binding:"required,oneof=password key"`
	CredentialID int64      `json:"credential_id" binding:"required"`
	Tags         model.Tags `json:"tags"`
}

type UpdateReq struct {
	Name         string     `json:"name" binding:"required,max=100"`
	IP           string     `json:"ip" binding:"required"`
	Port         int        `json:"port"`
	AuthType     string     `json:"auth_type" binding:"required,oneof=password key"`
	CredentialID int64      `json:"credential_id" binding:"required"`
	Tags         model.Tags `json:"tags"`
}

func (s *Service) Create(tenantID int64, req CreateReq) (*model.Host, error) {
	port := req.Port
	if port == 0 {
		port = 22
	}
	h := &model.Host{
		TenantID:     tenantID,
		Name:         req.Name,
		IP:           req.IP,
		Port:         port,
		AuthType:     req.AuthType,
		CredentialID: req.CredentialID,
		Status:       1,
		Tags:         req.Tags,
	}
	if err := s.repo.Create(h); err != nil {
		return nil, fmt.Errorf("create host: %w", err)
	}
	return h, nil
}

func (s *Service) Update(tenantID, id int64, req UpdateReq) (*model.Host, error) {
	h, err := s.repo.GetByID(tenantID, id)
	if err != nil {
		return nil, fmt.Errorf("host not found")
	}
	h.Name = req.Name
	h.IP = req.IP
	h.Port = req.Port
	if h.Port == 0 {
		h.Port = 22
	}
	h.AuthType = req.AuthType
	h.CredentialID = req.CredentialID
	h.Tags = req.Tags
	if err := s.repo.Update(h); err != nil {
		return nil, fmt.Errorf("update host: %w", err)
	}
	return h, nil
}

func (s *Service) Delete(tenantID, id int64) error {
	return s.repo.Delete(tenantID, id)
}

func (s *Service) List(tenantID int64, page, pageSize int) ([]model.Host, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	return s.repo.List(tenantID, page, pageSize)
}

func (s *Service) GetByID(tenantID, id int64) (*model.Host, error) {
	return s.repo.GetByID(tenantID, id)
}
