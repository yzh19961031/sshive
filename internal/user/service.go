package user

import (
	"fmt"

	"github.com/sshive/sshive/internal/auth"
	"github.com/sshive/sshive/internal/model"
)

type Service struct {
	repo *Repo
}

func NewService() *Service {
	return &Service{repo: &Repo{}}
}

type CreateReq struct {
	Username string `json:"username" binding:"required,min=2,max=50"`
	Password string `json:"password" binding:"required,min=8"`
	Email    string `json:"email"`
}

type AssignRolesReq struct {
	RoleIDs []int64 `json:"role_ids" binding:"required"`
}

func (s *Service) Create(tenantID int64, req CreateReq) (*model.User, error) {
	hash, err := auth.HashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}
	u := &model.User{
		TenantID:     tenantID,
		Username:     req.Username,
		PasswordHash: hash,
		Email:        req.Email,
		Status:       1,
	}
	if err := s.repo.Create(u); err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}
	return u, nil
}

func (s *Service) List(tenantID int64, page, pageSize int) ([]model.User, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	return s.repo.List(tenantID, page, pageSize)
}

func (s *Service) AssignRoles(tenantID, userID int64, req AssignRolesReq) error {
	if _, err := s.repo.GetByID(tenantID, userID); err != nil {
		return fmt.Errorf("user not found")
	}
	return s.repo.SetRoles(userID, req.RoleIDs)
}
