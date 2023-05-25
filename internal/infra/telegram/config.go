package telegram

type Config struct {
	Token string `env:"TELEGRAM_TOKEN,required"`
	Debug bool   `env:"TELEGRAM_DEBUG,default=false"`
}
