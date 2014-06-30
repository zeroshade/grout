// Copyright (C) 2014 zeroshade. All rights reserved
// Use of this source code is goverened by the GPLv2 license
// which can be found in the license.txt file

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
		PrintFPS     bool `gcfg:"printfps"`
		ShowSprBound bool `gcfg:"showspritebounds"`
	}
	Paths struct {
		Res string `gcfg:"resources"`
		Spr string `gcfg:"sprites"`
	}
}

func loadSettings(c *Config) error {
	return gcfg.ReadFileInto(c, "settings.ini")
}
