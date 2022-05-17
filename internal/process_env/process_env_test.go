package processenv

import (
	"os"
	"strings"
	"testing"
	"time"

	"github.com/inhuman/pinger/internal/checks"
	"github.com/inhuman/tools/app"
)

func TestParseEnv(t *testing.T) {
	cases := []struct {
		args   []string
		result map[string]checks.CheckParam
	}{
		{
			args: []string{
				"HOST_EXAMPLE=http://example.com",
				"PERIOD_EXAMPLE=3m",
				"LATENCY_EXAMPLE=2m",
				"HOST_TEST2=exampleexample.com",
			},
			result: map[string]checks.CheckParam{
				"http://example.com": {Period: 3 * time.Minute, Latency: 2 * time.Minute},
				"exampleexample.com": {
					Period: 10 * time.Minute, Latency: 1 * time.Second,
				},
			},
		},
	}

	for i := range cases {
		// set envs
		for _, e := range cases[i].args {
			key := strings.Split(e, "=")[0]
			value := strings.Split(e, "=")[1]

			err := os.Setenv(key, value)
			if err != nil {
				t.Fatal(err)
			}
		}

		// parse envs
		res := parseEnv(app.InitLogger(), cases[i].args)

		// check envs
		if len(res) != len(cases[i].result) {
			t.Fatalf(
				"map length mismatch: expected %d, but got %d",
				len(cases[i].result),
				len(res),
			)
		}

		for key, value := range cases[i].result {
			v, isExists := res[key]
			if !isExists {
				t.Fatalf("key %q has not been parsed", key)
			}

			if v.Latency != value.Latency {
				t.Errorf(
					"Latency of %q mismatch: expected %v, but got %v",
					key,
					value.Latency,
					v.Latency,
				)
			}

			if v.Period != value.Period {
				t.Errorf(
					"Period of %q mismatch: expected %v, but got %v",
					key,
					value.Period,
					v.Period,
				)
			}
		}

		// unset envs
		for _, e := range cases[i].args {
			key := strings.Split(e, "=")[0]

			err := os.Unsetenv(key)
			if err != nil {
				t.Fatal(err)
			}
		}
	}
}
