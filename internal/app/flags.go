package app

import (
	"flag"

	"github.com/caarlos0/env"

	"github.com/Stern-Ritter/gophermart/internal/config"
)

func GetConfig(c config.ServerConfig) (config.ServerConfig, error) {
	err := parseFlags(&c)
	if err != nil {
		return c, err
	}

	err = env.Parse(&c)
	if err != nil {
		return c, err
	}
	err = env.Parse(&c.ProcessAccrualsConfig)

	return c, err
}

func parseFlags(c *config.ServerConfig) error {
	flag.StringVar(&c.URL, "a", ":8080", "address to run gophermart in format <host>:<port>")
	flag.StringVar(&c.DatabaseURL, "d", "", "database URL")
	flag.StringVar(&c.AccrualSystemURL, "r", "", "address for sending requests to loyalty point accrual system")
	flag.StringVar(&c.JwtSecretKey, "k", "secretKey", "secret used for jwt key")
	flag.IntVar(&c.ProcessAccrualsConfig.ProcessAccrualsBatchMaxSize, "bs", 1, "processing accruals batch max size")
	flag.IntVar(&c.ProcessAccrualsConfig.ProcessAccrualsBufferSize, "s", 10, "processing accruals buffer size")
	flag.IntVar(&c.ProcessAccrualsConfig.ProcessAccrualsWorkerPoolSize, "w", 10, "processing accruals worker pool size")
	flag.IntVar(&c.ProcessAccrualsConfig.GetNewAccrualsInterval, "i", 1, "interval to fetch new accruals")

	return nil
}
