package main

import (
	"context"
	"log"
	"strings"
	"time"

	"github.com/go-logr/logr"
	"github.com/joho/godotenv"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/inhuman/pinger"
	"github.com/inhuman/pinger/internal/checks"
	processenv "github.com/inhuman/pinger/internal/process_env"
	"github.com/inhuman/tools/app"
	"github.com/inhuman/tools/processes"
)

const (
	metricsPostfix = "_availability"
)

type checkList map[string]checks.CheckParam

func main() {
	if godotenv.Load() != nil {
		log.Println("envs loaded from OS")
	} else {
		log.Println("env loaded from files")
	}

	appInstance := app.New(pinger.ServiceName, app.WithPrometheus())

	defer app.RecoverExit(appInstance.Logr())

	checkList := processenv.ReadEnv(appInstance.Logr())

	workers, metrics := getWorkers(checkList, appInstance.Logr())

	collectors := make([]prometheus.Collector, 0, len(metrics))

	for i := range metrics {
		collectors = append(collectors, metrics[i])
	}

	appInstance.MustPromRegister(collectors...)

	appInstance.AddWorkRunners(
		workers...,
	)

	appInstance.Run()
}

func worker(ctx context.Context, link string, params checks.CheckParam, infoMetrics *prometheus.GaugeVec,
	logger logr.Logger,
) {
	ticker := time.NewTicker(params.Period)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			err := checks.CheckWithTimeout(logger, link, params)
			if err != nil {
				infoMetrics.WithLabelValues(link, err.Error()).Set(1)
			} else {
				infoMetrics.WithLabelValues(link, "ok").Set(1)
			}
		case <-ctx.Done():
			return
		}
	}
}

func getWorkers(checkList checkList, logger logr.Logger) ([]processes.Process, []*prometheus.GaugeVec) {
	metrics := make([]*prometheus.GaugeVec, 0, len(checkList))
	workers := make([]processes.Process, 0, len(checkList))

	for link := range checkList {
		params := checkList[link]

		func(link string, params checks.CheckParam) {
			metric := prometheus.NewGaugeVec(prometheus.GaugeOpts{
				Name: strings.ToLower(params.EnvName) + metricsPostfix,
			},
				[]string{"address", "status"},
			)

			metrics = append(metrics, metric)
			workers = append(workers, func(ctx context.Context) error {
				worker(ctx, link, params, metric, logger)

				return nil
			})
		}(link, params)
	}

	return workers, metrics
}
