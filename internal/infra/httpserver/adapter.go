package httpserver

import (
	"context"
	"errors"
	"github.com/L11R/masked-email-bot/internal/domain"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"
	"net/http"
)

type delivery struct {
	logger  *zap.Logger
	config  *Config
	service domain.Service

	server *http.Server
}

func NewDelivery(logger *zap.Logger, config *Config, service domain.Service) (domain.Delivery, error) {
	a := &delivery{
		logger:  logger,
		config:  config,
		service: service,
	}

	r := chi.NewRouter()
	r.Use(middleware.RequestLogger(&logFormatter{
		logger: logger,
	}))
	r.Get("/redirect", a.handleOAuth2Redirect)

	a.server = &http.Server{
		Addr:    config.Address,
		Handler: r,
	}

	return a, nil
}

// ListenAndServe HTTP requests.
func (d *delivery) ListenAndServe() error {
	d.logger.Info("Listening and serving HTTP requests.", zap.String("address", d.config.Address))

	if err := d.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		//if err := d.server.ListenAndServeTLS("cert.pem", "key.pem"); err != nil && !errors.Is(err, http.ErrServerClosed) {
		d.logger.Error("Error listening and serving HTTP requests!", zap.Error(err))
		return domain.ErrHTTPInternal
	}

	return nil
}

// Shutdown the HTTP adapter
func (d *delivery) Shutdown(ctx context.Context) error {
	if err := d.server.Shutdown(ctx); err != nil {
		d.logger.Error("Error shutting down HTTP adapter!", zap.Error(err))
		return domain.ErrHTTPInternal
	}

	return nil
}
