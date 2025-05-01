package config

import (
	"log"
	"os"

	"github.com/echo-webkom/cenv"
)

type Config struct {
	Port    string
	DumpDir string
}

func Load() *Config {
	if err := cenv.Load(); err != nil {
		log.Fatal(err)
	}

	return &Config{
		Port:    toGoPort(os.Getenv("PORT")),
		DumpDir: os.Getenv("DUMP_DIR"),
	}
}

func toGoPort(port string) string {
	if port[0] != ':' {
		return ":" + port
	}
	return port
}
