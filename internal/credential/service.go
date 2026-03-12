package credential

import (
	"encoding/base64"
	"fmt"

	"github.com/sshive/sshive/internal/config"
	"github.com/sshive/sshive/internal/model"
	"github.com/sshive/sshive/pkg/encrypt"
)

type Service struct {
	repo *Repo
}

func NewService() *Service {
	return &Service{repo: &Repo{}}
}

type CreateReq struct {
	Name     string `json:"name" binding:"required,max=100"`
	Type     string `json:"type" binding:"required,oneof=password key"`
	Username string `json:"username" binding:"required,max=50"`
	Secret   string `json:"secret" binding:"required"`
}

func (s *Service) encryptionKey() ([]byte, error) {
	return base64.StdEncoding.DecodeString(config.C.EncryptKey)
}

func (s *Service) Create(tenantID int64, req CreateReq) (*model.Credential, error) {
	key, err := s.encryptionKey()
	if err != nil {
		return nil, fmt.Errorf("load encrypt key: %w", err)
	}
	encrypted, err := encrypt.Encrypt(key, req.Secret)
	if err != nil {
		return nil, fmt.Errorf("encrypt secret: %w", err)
	}
	c := &model.Credential{
		TenantID:        tenantID,
		Name:            req.Name,
		Type:            req.Type,
		Username:        req.Username,
		EncryptedSecret: encrypted,
	}
	if err := s.repo.Create(c); err != nil {
		return nil, fmt.Errorf("save credential: %w", err)
	}
	return c, nil
}

func (s *Service) List(tenantID int64, page, pageSize int) ([]model.Credential, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	return s.repo.List(tenantID, page, pageSize)
}

// DecryptSecret 解密指定凭据的 secret，供 SSH/SFTP 模块调用
func (s *Service) DecryptSecret(tenantID, id int64) (username, secret string, err error) {
	c, err := s.repo.GetByID(tenantID, id)
	if err != nil {
		return "", "", fmt.Errorf("credential not found: %w", err)
	}
	key, err := s.encryptionKey()
	if err != nil {
		return "", "", fmt.Errorf("load encrypt key: %w", err)
	}
	plain, err := encrypt.Decrypt(key, c.EncryptedSecret)
	if err != nil {
		return "", "", fmt.Errorf("decrypt: %w", err)
	}
	return c.Username, plain, nil
}
