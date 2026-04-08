package config

import (
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Env           string `yaml:"env" env-default:"local"`
	MigrationPath string `yaml:"migration_path" env-default:"./migrations"`
	HTTPServer    `yaml:"http_server"`
	DBConfig      `yaml:"db_config"`
}

type HTTPServer struct {
	Address string        `yaml:"address" env-default:"localhost:8080"`
	Timeout time.Duration `yaml:"timeout" env-default:"4s"`
}

type DBConfig struct {
	Host     string `yaml:"host" env-required:"true" env:"DB_HOST"`
	Port     string `yaml:"port" env-required:"true" env:"DB_PORT"`
	User     string `yaml:"user" env-required:"true" env:"DB_USER"`
	Password string `yaml:"password" env-required:"true" env:"DB_PASSWORD"`
	Name     string `yaml:"name" env-required:"true" env:"DB_NAME"`
	SSLMode  string `yaml:"SSLMode" env-required:"true" env:"DB_SSLMODE"`
}

func MustLoad() *Config {
	_ = godotenv.Load()

	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		log.Fatal("CONFIG_PATH is not set")
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Fatalf("config file does not exist: %s", configPath)
	}

	var cfg Config

	if err := readConfig(configPath, &cfg); err != nil {
		log.Fatalf("cannot read config: %s", err)
	}

	return &cfg
}

func readConfig(path string, cfg interface{}) error {
	f, err := os.OpenFile(path, os.O_RDONLY|os.O_SYNC, 0)
	if err != nil {
		return err
	}
	defer f.Close()

	if err = yaml.NewDecoder(f).Decode(cfg); err != nil {
		return err
	}

	return readEnvVars(cfg, false)
}
