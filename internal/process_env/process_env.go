package processenv

import (
	"os"
	"strings"
	"time"

	"github.com/go-logr/logr"
	"github.com/inhuman/pinger/internal/checks"
	"github.com/inhuman/tools/env"
)

const (
	hostPrefix    = "HOST_"
	latencyPrefix = "LATENCY_"
	periodPrefix  = "PERIOD_"
)

var (
	defaultLatency = 1 * time.Second
	defaultPeriod  = 10 * time.Second
)

func ReadEnv(log logr.Logger) map[string]checks.CheckParam {
	return parseEnv(log, os.Environ())
}

func parseEnv(log logr.Logger, envVars []string) map[string]checks.CheckParam {
	checkList := make(map[string]checks.CheckParam)

	// get links values
	for i := range envVars {
		if strings.HasPrefix(envVars[i], hostPrefix) {
			// name in env variable
			varName := strings.TrimPrefix(strings.Split(envVars[i], "=")[0], hostPrefix)
			// save link
			link := env.Get(strings.Split(envVars[i], "=")[0]).MustString()

			checkList[link] = checks.CheckParam{
				Latency: defaultLatency,
				Period:  defaultPeriod,
				EnvName: varName,
			}
		}
	}

	// output links in stdout
	links := make([]string, 0, len(checkList))

	for link := range checkList {
		links = append(links, link)
	}

	log.Info("watching: " + strings.Join(links, ", "))

	// get latency values
	for link := range checkList {
		params := checkList[link]

		for i := range envVars {
			varName := strings.Split(envVars[i], "=")[0]

			if varName == latencyPrefix+checkList[link].EnvName {
				params.Latency = env.Get(varName).MustDuration()
				checkList[link] = params

				break
			}
		}
	}

	// get period values
	for link := range checkList {
		params := checkList[link]

		for i := range envVars {
			varName := strings.Split(envVars[i], "=")[0]

			if varName == periodPrefix+checkList[link].EnvName {
				params.Period = env.Get(varName).MustDuration()
				checkList[link] = params

				break
			}
		}
	}

	return checkList
}
