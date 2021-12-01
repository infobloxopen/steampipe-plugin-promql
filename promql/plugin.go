package promql

import (
	"context"
	"io/ioutil"
	"os"
	"time"

	papi "github.com/prometheus/client_golang/api"
	v1api "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
	"github.com/turbot/steampipe-plugin-sdk/plugin"
	"github.com/turbot/steampipe-plugin-sdk/plugin/transform"
	"gopkg.in/yaml.v2"
)

// Plugin returns a promql plugin for steampipe.
func Plugin(ctx context.Context) *plugin.Plugin {
	p := &plugin.Plugin{
		Name:             "steampipe-plugin-promql",
		DefaultTransform: transform.FromGo().NullIfZero(),
		TableMap: map[string]*plugin.Table{
			"promql_metric": tablePromqlMetric(),
			"promql_alias":  tablePromqlAlias(),
		},
	}
	return p
}

type Sample struct {
	Name      string
	Timestamp time.Time
	Value     float64
}

func getConfig() (*Config, error) {

	bs, err := ioutil.ReadFile(os.Getenv("HOME") + "/.steampipe-promql/config.yaml")
	if err != nil {
		return nil, err
	}
	var cfg Config
	if err := yaml.Unmarshal(bs, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

type PromClient interface {
	QueryRange(context.Context, string, v1api.Range) (model.Value, v1api.Warnings, error)
}

type Config struct {
	Endpoint string            `json:"endpoint"`
	Aliases  map[string]string `json:"aliases"`
}

func getAliasDef(name string) (string, bool) {
	cfg, err := getConfig()
	if err != nil {
		return "", false
	}
	a, found := cfg.Aliases[name]
	return a, found
}

func getPromClient(ctx context.Context) (PromClient, error) {

	cfg, err := getConfig()
	if err != nil {
		return nil, err
	}

	aclient, err := papi.NewClient(papi.Config{Address: cfg.Endpoint})
	if err != nil {
		return nil, err
	}

	return v1api.NewAPI(aclient), nil
}
