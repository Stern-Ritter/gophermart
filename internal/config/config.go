package config

type ProcessAccrualsConfig struct {
	ProcessAccrualsBatchMaxSize   int `env:"PROCESS_ACCRUALS_BATCH_MAX_SIZE"`
	ProcessAccrualsBufferSize     int `env:"PROCESS_ACCRUALS_BUFFER_SIZE"`
	ProcessAccrualsWorkerPoolSize int `env:"PROCESS_ACCRUALS_WORKER_POOL_SIZE"`
	GetNewAccrualsInterval        int `env:"GET_NEW_ACCRUALS_INTERVAL"`
}

type ServerConfig struct {
	URL                   string `env:"RUN_ADDRESS"`
	DatabaseURL           string `env:"DATABASE_URI"`
	AccrualSystemURL      string `env:"ACCRUAL_SYSTEM_ADDRESS"`
	JwtSecretKey          string `env:"JWT_SECRET_KEY"`
	ProcessAccrualsConfig ProcessAccrualsConfig
	LoggerLvl             string
}
