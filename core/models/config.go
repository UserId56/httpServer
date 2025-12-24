package models

type Config struct {
	Port            int
	Address         string
	MaxIdleConns    int
	MaxOpenConns    int
	ConnMaxLifetime int
	ConnMaxIdleTime int
}

func (c *Config) Init() {
	if c.Port == 0 {
		c.Port = 3000
	}
	if c.Address == "" {
		c.Address = ""
	}
	if c.MaxIdleConns == 0 {
		c.MaxIdleConns = 10
	}
	if c.MaxOpenConns == 0 {
		c.MaxOpenConns = 80
	}
	if c.ConnMaxLifetime == 0 {
		c.ConnMaxLifetime = 30
	}
	if c.ConnMaxIdleTime == 0 {
		c.ConnMaxIdleTime = 5
	}
}
