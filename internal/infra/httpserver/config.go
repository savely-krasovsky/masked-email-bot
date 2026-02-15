package httpserver

type Config struct {
	Address string `env:"HTTP_ADDRESS,default=0.0.0.0=8080"`
}
