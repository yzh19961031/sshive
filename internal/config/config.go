// internal/config/config.go
package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	DB struct {
		DSN string `yaml:"dsn"`
	} `yaml:"db"`

	Redis struct {
		Addr     string `yaml:"addr"`
		Password string `yaml:"password"`
	} `yaml:"redis"`

	EncryptKey string `yaml:"encrypt_key"`
	CastDir    string `yaml:"cast_dir"`
	JWTSecret  string `yaml:"jwt_secret"`
	Port       int    `yaml:"port"`
}

var C Config

func Load(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("open config file %q: %w", path, err)
	}
	defer f.Close()

	if err := yaml.NewDecoder(f).Decode(&C); err != nil {
		return fmt.Errorf("decode config file: %w", err)
	}

	if C.DB.DSN == "" {
		return fmt.Errorf("db.dsn is required")
	}
	if C.Redis.Addr == "" {
		return fmt.Errorf("redis.addr is required")
	}
	if C.EncryptKey == "" {
		return fmt.Errorf("encrypt_key is required")
	}
	if C.JWTSecret == "" {
		return fmt.Errorf("jwt_secret is required")
	}
	if C.CastDir == "" {
		C.CastDir = "/data/sshive/cast"
	}
	if C.Port == 0 {
		C.Port = 8080
	}
	return nil
}
