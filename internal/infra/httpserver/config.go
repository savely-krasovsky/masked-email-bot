package httpserver

type Config struct {
	Address string `env:"HTTP_ADDRESS,required"`
}
