package main

import (
	"github.com/joeshaw/envdecode"
	"github.com/prometheus/client_golang/api"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"

	"context"
	"log"
	"time"
)

const (
	errInvalidResultType = "Invalid result from query - returned model.Value must be type Matrix"
)

type Config struct {
	PrometheusEndpoint string        `env:"PROMETHEUS_ENDPOINT,default=localhost"`
	RangeQuery         string        `env:"RANGE_QUERY,required"`
	RangeTime          time.Duration `env:"RANGE_TIME,default=-5m"`
	TargetValue        int           `env:"TARGET_VALUE,required"`
	TargetStrategy     string        `env:"TARGET_STRATEGY,default=min"` // valid values: max, min, equals
	Timeout            time.Duration `env:"TIMEOUT,default=10m"`
	TickTime           time.Duration `env:"TICK_TIME,default=1m"`
}

func main() {
	log.Println("at=gate-start")

	var cfg Config
	if err := envdecode.Decode(&cfg); err != nil {
		log.Fatalf("Error decoding environment: %s\n", err)
	}
	log.Printf("TARGET_VALUE=%d TARGET_STRATEGY=%q RANGE_TIME=%q", cfg.TargetValue, cfg.TargetStrategy, cfg.RangeTime)

	client, err := api.NewClient(api.Config{
		Address: cfg.PrometheusEndpoint,
	})
	if err != nil {
		log.Fatalf("Error creating client: %s\n", err)
	}
	v1api := v1.NewAPI(client)

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	r := v1.Range{
		Start: time.Now().Add(cfg.RangeTime),
		End:   time.Now(),
		Step:  time.Minute,
	}

	ticker := time.NewTicker(cfg.TickTime)
Done:
	for {
		select {
		// timeout
		case <-time.After(time.Hour):
			log.Fatal("at=timeout-exceeded")
		case <-ticker.C:
			log.Println("at=querying-prometheus")
			if queryPrometheusState(ctx, v1api, cfg, r) {
				log.Println("at=success-condition-met")
				break Done
			}
		}
	}

	log.Println("at=gate-success")
}

func queryPrometheusState(ctx context.Context, v1api v1.API, cfg Config, r v1.Range) bool {
	result, warnings, err := v1api.QueryRange(ctx, cfg.RangeQuery, r)
	if err != nil {
		log.Printf("at=error-querying-prometheus error=%s", err)
		return false
	}
	if len(warnings) > 0 {
		log.Println("warnings: %s", warnings)
	}

	switch {
	case result.Type() == model.ValMatrix:
		val := result.(model.Matrix)
		if len(val) == 0 {
			log.Printf("at=empty-value-set returned len=%d", len(val))
			return false
		}

		for _, val := range val {
			if len(val.Values) == 0 {
				log.Printf("at=no-values-returned-for-range len=%d", len(val.Values))
				return false
			}

			for _, v := range val.Values {
				log.Printf("at=evaluating-value timestamp=%q value=%q", v.Timestamp, v.Value)
				switch cfg.TargetStrategy {
				case "min":
					if cfg.TargetValue > int(v.Value) {
						log.Printf("at=below-minimum min=%d value=%d", cfg.TargetValue, int(v.Value))
						return false

					}
				case "max":
					if cfg.TargetValue < int(v.Value) {
						log.Printf("at=above-max max=%d value=%d", cfg.TargetValue, int(v.Value))
						return false
					}
				case "equals":
					if cfg.TargetValue != int(v.Value) {
						log.Printf("at=unequals equals=%d value=%d", cfg.TargetValue, int(v.Value))
						return false
					}
				default:
					// just quit
					log.Fatalf("at=invalid-strategy strategy=%q", cfg.TargetStrategy)
				}
			}
		}
	default:
		log.Fatalf("at=invalid-query-result-type type=%s message=%q", result.Type(), errInvalidResultType)
	}

	return true
}
