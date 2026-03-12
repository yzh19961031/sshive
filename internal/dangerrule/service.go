package dangerrule

import (
	"fmt"
	"regexp"

	"github.com/sshive/sshive/internal/model"
)

type Service struct {
	repo *Repo
}

func NewService() *Service {
	return &Service{repo: &Repo{}}
}

type CreateReq struct {
	Pattern     string `json:"pattern" binding:"required,max=500"`
	Description string `json:"description" binding:"max=255"`
	Action      string `json:"action" binding:"omitempty,oneof=block warn"`
}

type UpdateReq struct {
	Pattern     string `json:"pattern" binding:"required,max=500"`
	Description string `json:"description" binding:"max=255"`
	Action      string `json:"action" binding:"omitempty,oneof=block warn"`
	Enabled     *int8  `json:"enabled"`
}

func (s *Service) Create(tenantID int64, req CreateReq) (*model.DangerRule, error) {
	if _, err := regexp.Compile(req.Pattern); err != nil {
		return nil, fmt.Errorf("invalid regex: %w", err)
	}
	action := req.Action
	if action == "" {
		action = "block"
	}
	rule := &model.DangerRule{
		TenantID:    tenantID,
		Pattern:     req.Pattern,
		Description: req.Description,
		Action:      action,
		Enabled:     1,
	}
	if err := s.repo.Create(rule); err != nil {
		return nil, fmt.Errorf("create rule: %w", err)
	}
	// 刷新匹配器缓存
	_ = GlobalMatcher.Reload(tenantID)
	return rule, nil
}

func (s *Service) Update(tenantID, id int64, req UpdateReq) (*model.DangerRule, error) {
	rules, _, err := s.repo.List(tenantID, 1, 1000)
	if err != nil {
		return nil, err
	}
	var rule *model.DangerRule
	for i := range rules {
		if rules[i].ID == id {
			rule = &rules[i]
			break
		}
	}
	if rule == nil {
		return nil, fmt.Errorf("rule not found")
	}
	if _, err := regexp.Compile(req.Pattern); err != nil {
		return nil, fmt.Errorf("invalid regex: %w", err)
	}
	rule.Pattern = req.Pattern
	rule.Description = req.Description
	if req.Action != "" {
		rule.Action = req.Action
	}
	if req.Enabled != nil {
		rule.Enabled = *req.Enabled
	}
	if err := s.repo.Update(rule); err != nil {
		return nil, fmt.Errorf("update rule: %w", err)
	}
	_ = GlobalMatcher.Reload(tenantID)
	return rule, nil
}

func (s *Service) Delete(tenantID, id int64) error {
	if err := s.repo.Delete(tenantID, id); err != nil {
		return err
	}
	_ = GlobalMatcher.Reload(tenantID)
	return nil
}

func (s *Service) List(tenantID int64, page, pageSize int) ([]model.DangerRule, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	return s.repo.List(tenantID, page, pageSize)
}
