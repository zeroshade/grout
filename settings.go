package grout

import (
	"code.google.com/p/gcfg"
	"time"
)

type Config struct {
	Video struct {
		W   uint          `gcfg:"width"`
		H   uint          `gcfg:"height"`
		FPS time.Duration `gcfg:"fps"`
	}
	Debug struct {
		PrintFPS bool `gcfg:"printfps"`
	}
	Paths struct {
		Res string `gcfg:"resources"`
		Spr string `gcfg:"sprites"`
	}
}

func loadSettings(c *Config) error {
	return gcfg.ReadFileInto(c, "settings.ini")
}
