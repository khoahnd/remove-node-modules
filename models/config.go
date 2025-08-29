package models

type Config struct {
	RootPath string
	Workers  int
	DryRun   bool
	ShowHelp bool
}

func NewConfig(rootPath string, workers int, dryRun bool) *Config {
	return &Config{
		RootPath: rootPath,
		Workers:  workers,
		DryRun:   dryRun,
	}
}

func (c *Config) IsValid() bool {
	return c.RootPath != "" && c.Workers > 0
}
