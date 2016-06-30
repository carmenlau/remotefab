package config

import (
  "os"
  "strings"
)

// Configuration - Remote fabric config
type Configuration struct {
	WorkingDirPath string
  Port         string
}

// NewConfigFromEnv - New config instance from enviroment variable
func NewConfigFromEnv() Configuration {
  c := Configuration{
    WorkingDirPath: "/tmp/",
    Port: ":8080",
  }
  if os.Getenv("REMOTE_FABRIC_CLONE_DIR_PATH") != "" {
    c.WorkingDirPath = os.Getenv("REMOTE_FABRIC_CLONE_DIR_PATH")
  }

  if os.Getenv("REMOTE_FABRIC_PORT") != "" {
    p := os.Getenv("REMOTE_FABRIC_PORT")
    if !strings.HasPrefix(p, ":") {
      c.Port = strings.Join([]string{":", p}, "")
    } else {
      c.Port = p
    }
  }

  return c
}
