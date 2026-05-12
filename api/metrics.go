package api

import (
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

func (m *MetricsServer) Run() error {
	m.echo.HideBanner = true
	m.echo.GET("/metrics", echo.WrapHandler(promhttp.Handler()))
	m.logger.Info("metrics listening on", zap.String("address", m.address))
	return m.echo.Start(m.address)
}
