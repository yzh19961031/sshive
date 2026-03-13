package db

import (
	"log/slog"

	"github.com/sshive/sshive/internal/model"
	"golang.org/x/crypto/bcrypt"
)

// Seed 初始化默认数据（幂等，可重复执行）
func Seed() error {
	// 1. 默认租户
	tenant := model.Tenant{Name: "默认租户", Status: 1}
	if err := DB.Where("name = ?", "默认租户").FirstOrCreate(&tenant).Error; err != nil {
		return err
	}
	slog.Info("seed: tenant ready", "id", tenant.ID, "name", tenant.Name)

	// 2. 权限列表
	perms := []struct {
		code string
		desc string
	}{
		{"host:connect", "连接主机（SSH/SFTP）"},
		{"host:manage", "管理主机和凭据"},
		{"sftp:access", "SFTP 文件管理"},
		{"audit:view", "查看审计日志"},
		{"user:manage", "管理用户和角色"},
		{"rule:manage", "管理高危规则"},
		{"tenant:manage", "管理租户"},
	}
	permMap := make(map[string]model.Permission)
	for _, p := range perms {
		var perm model.Permission
		if err := DB.Where("code = ?", p.code).FirstOrCreate(&perm, model.Permission{
			Code:        p.code,
			Description: p.desc,
		}).Error; err != nil {
			return err
		}
		permMap[p.code] = perm
	}

	// 3. 角色 + 角色权限
	roles := []struct {
		name  string
		desc  string
		perms []string
	}{
		{"super_admin", "超级管理员", []string{
			"host:connect", "host:manage", "sftp:access",
			"audit:view", "user:manage", "rule:manage", "tenant:manage",
		}},
		{"tenant_admin", "租户管理员", []string{
			"host:connect", "host:manage", "sftp:access",
			"audit:view", "user:manage", "rule:manage",
		}},
		{"operator", "运维人员", []string{
			"host:connect", "sftp:access", "audit:view",
		}},
		{"auditor", "审计人员", []string{
			"audit:view",
		}},
	}

	roleMap := make(map[string]model.Role)
	for _, r := range roles {
		var role model.Role
		if err := DB.Where("tenant_id = ? AND name = ?", tenant.ID, r.name).
			FirstOrCreate(&role, model.Role{
				TenantID:    tenant.ID,
				Name:        r.name,
				Description: r.desc,
			}).Error; err != nil {
			return err
		}
		roleMap[r.name] = role

		// 角色权限绑定
		for _, code := range r.perms {
			perm := permMap[code]
			var rp model.RolePermission
			DB.Where("role_id = ? AND permission_id = ?", role.ID, perm.ID).
				FirstOrCreate(&rp, model.RolePermission{
					RoleID:       role.ID,
					PermissionID: perm.ID,
				})
		}
	}

	// 4. 默认管理员用户 admin / Admin@123
	var admin model.User
	if err := DB.Where("tenant_id = ? AND username = ?", tenant.ID, "admin").
		First(&admin).Error; err != nil {
		hashBytes, hashErr := bcrypt.GenerateFromPassword([]byte("Admin@123"), bcrypt.DefaultCost)
		if hashErr != nil {
			return hashErr
		}
		admin = model.User{
			TenantID:     tenant.ID,
			Username:     "admin",
			PasswordHash: string(hashBytes),
			Email:        "admin@sshive.local",
			Status:       1,
		}
		if err := DB.Create(&admin).Error; err != nil {
			return err
		}
		slog.Info("seed: admin user created", "username", "admin", "password", "Admin@123")
	}

	// 5. 绑定 admin -> super_admin
	superAdminRole := roleMap["super_admin"]
	var ur model.UserRole
	DB.Where("user_id = ? AND role_id = ?", admin.ID, superAdminRole.ID).
		FirstOrCreate(&ur, model.UserRole{
			UserID: admin.ID,
			RoleID: superAdminRole.ID,
		})

	slog.Info("seed: initialization complete")
	return nil
}
