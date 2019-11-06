package config

import (
	"github.com/caarlos0/env"
	"github.com/go-playground/log"
	"github.com/joho/godotenv"
	_ "github.com/joho/godotenv/autoload"
)

type Config struct {
	AppName   string `env:"APP_NAME" envDefault:"featherr"`
	AppEnv    string `env:"APP_ENV" envDefault:"local"`
	AppDebug  string `env:"APP_DEBUG" envDefault:"true"`
	AppPort   string `env:"APP_PORT" envDefault:"8080"`
	AppDomain string `env:"APP_DOMAIN" envDefault:"0.0.0.0"`

	DBName string `env:"DB_NAME" envDefault:"feather_test"`
	DBPass string `env:"DB_PASS" envDefault:"Ibrahim22"`
	DBUser string `env:"DB_USER" envDefault:"root"`
	DBType string `env:"DB_TYPE" envDefault:"mysql"`
	DBHost string `env:"DB_HOST" envDefault:"172.17.0.1"`
	DBPort string `env:"DB_PORT" envDefault:"3306"`

	BucketName string `env:"BUCKET_NAME" envDefault:"featherr.appspot.com"`

	JWTSecret     string `env:"JWT_SECRET" envDefault:"qBPXnbcuQyauqlhTpJQjgAnmauKiZUgrhdu7eQhuNXfr6"`
	SpacesKey     string `env:"SPACES_KEY" envDefault:"6F7TJUERMUWIKSK5ECE6"`
	SpacesSecret  string `env:"SPACES_SECRET" envDefault:"lsVOQLkasuCIM3vltBe1b1XVnJeFHJfdt9brklP6zUU"`
	MailGunApiKey string `env:"MAILGUN_API_KEY" envDefault:""`
}

var cfg *Config

// Parse parses, validates and then returns the application
// configuration based on ENV variables
func init() {
	if err := godotenv.Load(".env"); err != nil {
		log.Warn("File .env not found, reading configuration from ENV")
	}

	log.Info("Parsing ENV vars...")
	defer log.Info("Done Parsing ENV vars")

	cfg = &Config{}

	if err := env.Parse(cfg); err != nil {
		log.WithFields(log.F("error", err)).Warn("Errors Parsing Configuration")
	}

	return
}

func GetConfig() *Config {
	return cfg
}
