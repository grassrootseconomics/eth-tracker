package api

import (
	"net/http"

	"github.com/VictoriaMetrics/metrics"
	"github.com/grassrootseconomics/celo-tracker/internal/stats"
	"github.com/uptrace/bunrouter"
)

func New(statsCollector *stats.Stats) *bunrouter.Router {
	router := bunrouter.New()

	router.GET("/metrics", metricsHandler())
	router.GET("/stats", statsHandler(statsCollector))

	return router
}

func metricsHandler() bunrouter.HandlerFunc {
	return func(w http.ResponseWriter, _ bunrouter.Request) error {
		metrics.WritePrometheus(w, true)
		return nil
	}
}

func statsHandler(s *stats.Stats) bunrouter.HandlerFunc {
	return func(w http.ResponseWriter, _ bunrouter.Request) error {
		return bunrouter.JSON(w, s.APIStatsResponse())
	}
}
