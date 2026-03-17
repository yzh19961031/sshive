// internal/dbclient/service.go
package dbclient

import (
	"context"
	"database/sql"
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	"github.com/sshive/sshive/internal/config"
	"github.com/sshive/sshive/internal/model"
	"github.com/sshive/sshive/pkg/encrypt"
)

const (
	queryTimeout = 30 * time.Second
	maxRows      = 1000

	DBTypeMySQL    = "mysql"
	DBTypePostgres = "postgres"
)

type QueryResult struct {
	Columns []string        `json:"columns"`
	Rows    [][]interface{} `json:"rows"`
	Total   int             `json:"total"`
}

type Service struct {
	repo *Repo
}

func NewService(repo *Repo) *Service {
	return &Service{repo: repo}
}

func (s *Service) encryptionKey() ([]byte, error) {
	return base64.StdEncoding.DecodeString(config.C.EncryptKey)
}

func (s *Service) openDB(srv *model.DBServer) (*sql.DB, error) {
	key, err := s.encryptionKey()
	if err != nil {
		return nil, fmt.Errorf("load encrypt key: %w", err)
	}
	pwd, err := encrypt.Decrypt(key, srv.EncryptedPassword)
	if err != nil {
		return nil, fmt.Errorf("decrypt password: %w", err)
	}

	var dsn, driver string
	switch srv.Type {
	case DBTypeMySQL:
		driver = DBTypeMySQL
		db := srv.Database
		if db == "" {
			db = "information_schema"
		}
		dsn = fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?timeout=5s&parseTime=true",
			srv.Username, pwd, srv.Host, srv.Port, db)
	case DBTypePostgres:
		driver = "postgres"
		dsn = fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable connect_timeout=5",
			srv.Host, srv.Port, srv.Username, pwd, srv.Database)
	default:
		return nil, fmt.Errorf("unsupported db type: %s", srv.Type)
	}

	db, err := sql.Open(driver, dsn)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	return db, nil
}

// Query 执行单条 SELECT，返回结果（超过 maxRows 截断）
func (s *Service) Query(tenantID, serverID int64, sqlStr string) (*QueryResult, error) {
	// 安全检查：禁止多语句
	if strings.Count(strings.TrimRight(sqlStr, "; \t\n"), ";") > 0 {
		return nil, fmt.Errorf("不支持多语句执行")
	}

	srv, err := s.repo.GetByID(tenantID, serverID)
	if err != nil {
		return nil, err
	}

	db, err := s.openDB(srv)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), queryTimeout)
	defer cancel()

	rows, err := db.QueryContext(ctx, sqlStr)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	cols, _ := rows.Columns()
	result := &QueryResult{Columns: cols}

	for rows.Next() {
		vals := make([]interface{}, len(cols))
		ptrs := make([]interface{}, len(cols))
		for i := range vals {
			ptrs[i] = &vals[i]
		}
		if err := rows.Scan(ptrs...); err != nil {
			continue
		}
		row := make([]interface{}, len(cols))
		for i, v := range vals {
			if b, ok := v.([]byte); ok {
				row[i] = string(b)
			} else {
				row[i] = v
			}
		}
		result.Rows = append(result.Rows, row)
		result.Total++
		if result.Total >= maxRows {
			break
		}
	}
	return result, rows.Err()
}

// ListDatabases 获取所有数据库名
func (s *Service) ListDatabases(tenantID, serverID int64) ([]string, error) {
	srv, err := s.repo.GetByID(tenantID, serverID)
	if err != nil {
		return nil, err
	}
	db, err := s.openDB(srv)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	var q string
	switch srv.Type {
	case DBTypeMySQL:
		q = "SHOW DATABASES"
	case DBTypePostgres:
		q = "SELECT datname FROM pg_database WHERE datistemplate = false"
	}
	rows, err := db.Query(q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var dbs []string
	for rows.Next() {
		var name string
		rows.Scan(&name)
		dbs = append(dbs, name)
	}
	return dbs, nil
}

// ListTables 获取指定数据库的表名
func (s *Service) ListTables(tenantID, serverID int64, database string) ([]string, error) {
	srv, err := s.repo.GetByID(tenantID, serverID)
	if err != nil {
		return nil, err
	}
	orig := srv.Database
	srv.Database = database
	db, err := s.openDB(srv)
	srv.Database = orig
	if err != nil {
		return nil, err
	}
	defer db.Close()

	var q string
	switch srv.Type {
	case DBTypeMySQL:
		q = "SHOW TABLES"
	case DBTypePostgres:
		q = "SELECT tablename FROM pg_tables WHERE schemaname='public'"
	}
	rows, err := db.Query(q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var name string
		rows.Scan(&name)
		tables = append(tables, name)
	}
	return tables, nil
}
