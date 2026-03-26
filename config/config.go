package config

import (
	"github.com/spf13/viper"
)

type Config struct {
	DBHost            string `mapstructure:"DB_HOST"`
	DBPort            string `mapstructure:"DB_PORT"`
	DBName            string `mapstructure:"DB_NAME"`
	DBUser            string `mapstructure:"DB_USER"`
	DBPassword        string `mapstructure:"DB_PASSWORD"`
	JWTSecret         string `mapstructure:"JWT_SECRET"`
	JWTExpiryMinutes  int    `mapstructure:"JWT_EXPIRY_MINUTES"`
	RefreshExpiryDays int    `mapstructure:"REFRESH_EXPIRY_DAYS"`
	AESKey            string `mapstructure:"AES_ENCRYPTION_KEY"`
}

// AESKeyBytes returns the AES key as exactly 32 bytes (padded or truncated).
func (c *Config) AESKeyBytes() []byte {
	key := []byte(c.AESKey)
	if len(key) >= 32 {
		return key[:32]
	}
	padded := make([]byte, 32)
	copy(padded, key)
	return padded
}

func Load() (*Config, error) {
	viper.SetConfigFile(".env")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	cfg := &Config{}
	if err := viper.Unmarshal(cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}
