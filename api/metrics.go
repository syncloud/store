package api

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

type MetricsServer struct {
	address string
	echo    *echo.Echo
	logger  *zap.Logger
}

func NewMetricsServer(address string, logger *zap.Logger) *MetricsServer {
	return &MetricsServer{
		address: address,
		echo:    echo.New(),
		logger:  logger,
	}
}

func (m *MetricsServer) Start() {
	m.echo.HideBanner = true
	m.echo.GET("/metrics", echo.WrapHandler(promhttp.Handler()))
	m.logger.Info("metrics listening on", zap.String("address", m.address))
	go func() {
		err := m.echo.Start(m.address)
		if err != nil && err != http.ErrServerClosed {
			m.logger.Error("metrics server stopped", zap.Error(err))
		}
	}()
}
