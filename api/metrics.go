package api

import (
	"github.com/labstack/echo/v4"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

type MetricsServer struct {
	address    string
	collectors []prometheus.Collector
	echo       *echo.Echo
	logger     *zap.Logger
}

func NewMetricsServer(address string, logger *zap.Logger, collectors ...prometheus.Collector) *MetricsServer {
	return &MetricsServer{
		address:    address,
		collectors: collectors,
		echo:       echo.New(),
		logger:     logger,
	}
}

func (m *MetricsServer) Start() <-chan error {
	registry := prometheus.NewRegistry()
	for _, c := range m.collectors {
		registry.MustRegister(c)
	}
	m.echo.HideBanner = true
	m.echo.GET("/metrics", echo.WrapHandler(promhttp.HandlerFor(registry, promhttp.HandlerOpts{})))
	m.logger.Info("metrics listening on", zap.String("address", m.address))
	errs := make(chan error, 1)
	go func() { errs <- m.echo.Start(m.address) }()
	return errs
}
