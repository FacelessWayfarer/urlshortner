package config

import (
	"errors"
	"io/fs"
	"log"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
)

type Cfg struct {
	Env          string `yaml:"env" env:"ENV" env-deafault:"local"`
	DatabasePath string `yaml:"database_path"  env-required:"true"`
	HTTPServ     `yaml:"http_server"`
}

type HTTPServ struct {
	Address     string        `yaml:"address"  env-deafault:":80"`
	Timeout     time.Duration `yaml:"timeout"  env-deafault:"4s"`
	IdleTimeout time.Duration `yaml:"idle_timeout"  env-deafault:"60s"`
	User        string        `yaml:"user" env-required:"true"`
	Password    string        `yaml:"password" env-required:"true" env:"HTTP_SERVER_PASSWORD"`
}

func MustLoad() *Cfg {
	godotenv.Load(".env")
	configPath := os.Getenv("CONFIG_PATH")
	log.Println(configPath)
	if configPath == "" {
		log.Fatal("CONFIG_PATH is not set")
	}

	if _, err := os.Stat(configPath); errors.Is(err, fs.ErrNotExist) {
		log.Fatalf("config file %s does not exist", configPath)
	}

	var cfg Cfg

	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		log.Fatalf("cannot read config: %s", err)
	}

	return &cfg
}
