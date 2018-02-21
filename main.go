package main

import (
	"easytracebot/easytraceCache"
	"github.com/go-kit/kit/log"
	"os"
	"github.com/go-kit/kit/log/level"
	kitprometheus "github.com/go-kit/kit/metrics/prometheus"
	stdprometheus "github.com/prometheus/client_golang/prometheus"
	"github.com/zamedic/go2hal/database"
	"github.com/zamedic/go2hal/telegram"
	"net/http"
	"os/signal"
	"syscall"
	"fmt"
	"github.com/zamedic/go2hal/chef"
	"github.com/zamedic/go2hal/alert"
)

func main() {
	var logger log.Logger
	logger = log.NewLogfmtLogger(log.NewSyncWriter(os.Stderr))
	logger = level.NewFilter(logger, level.AllowAll())
	logger = log.With(logger, "ts", log.DefaultTimestamp)
	fieldKeys := []string{"method"}

	db := database.NewConnection()
	telegramStore := telegram.NewMongoStore(db)
	chefStore := chef.NewMongoStore(db)
	alertStore := alert.NewStore(db)

	telegramService := telegram.NewService(telegramStore)

	alertService := alert.NewService(telegramService, alertStore)
	alertService = alert.NewLoggingService(log.With(logger, "component", "alert"), alertService)
	alertService = alert.NewInstrumentService(kitprometheus.NewCounterFrom(stdprometheus.CounterOpts{
		Namespace: "api",
		Subsystem: "alert_service",
		Name:      "request_count",
		Help:      "Number of requests received.",
	}, fieldKeys),
		kitprometheus.NewSummaryFrom(stdprometheus.SummaryOpts{
			Namespace: "api",
			Subsystem: "alert_service",
			Name:      "request_latency_microseconds",
			Help:      "Total duration of requests in microseconds.",
		}, fieldKeys), alertService)

	chefService := chef.NewService(alertService, chefStore)
	chefService = chef.NewLoggingService(log.With(logger, "component", "chef"), chefService)
	chefService = chef.NewInstrumentService(kitprometheus.NewCounterFrom(stdprometheus.CounterOpts{
		Namespace: "api",
		Subsystem: "chef",
		Name:      "request_count",
		Help:      "Number of requests received.",
	}, fieldKeys),
		kitprometheus.NewSummaryFrom(stdprometheus.SummaryOpts{
			Namespace: "api",
			Subsystem: "chef",
			Name:      "request_latency_microseconds",
			Help:      "Total duration of requests in microseconds.",
		}, fieldKeys), chefService)

	cacheService := easytraceCache.NewService(chefService)
	cacheService = easytraceCache.NewLoggingService(log.With(logger, "component", "easytraceCache"), cacheService)
	cacheService = easytraceCache.NewInstrumentService(kitprometheus.NewCounterFrom(stdprometheus.CounterOpts{
		Namespace: "api",
		Subsystem: "easytraceCache",
		Name:      "request_count",
		Help:      "Number of requests received.",
	}, fieldKeys),
		kitprometheus.NewSummaryFrom(stdprometheus.SummaryOpts{
			Namespace: "api",
			Subsystem: "easytraceCache",
			Name:      "request_latency_microseconds",
			Help:      "Total duration of requests in microseconds.",
		}, fieldKeys), cacheService)

	telegramService.RegisterCommand(easytraceCache.ReloadCacheCallCommand(telegramService, cacheService))

	errs := make(chan error, 2)

	go func() {
		logger.Log("transport", "http", "address", ":8000", "msg", "listening")
		errs <- http.ListenAndServe(":8000", nil)
	}()
	go func() {
		c := make(chan os.Signal)
		signal.Notify(c, syscall.SIGINT)
		errs <- fmt.Errorf("%s", <-c)
	}()

	logger.Log("terminated", <-errs)

}
