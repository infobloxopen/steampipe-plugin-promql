package promql

import (
	"context"
	"fmt"

	"github.com/turbot/steampipe-plugin-sdk/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/plugin"
	"github.com/turbot/steampipe-plugin-sdk/plugin/transform"
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

func tablePromqlMetric() *plugin.Table {
	return &plugin.Table{
		Name:        "promql_metric",
		Description: `Metric executes promql queries defined in the alias table. The queries are sumarized so no more than about 1000 data points come back for each query. For example, if you query for 30 days and the underlying data is sampled at 15 second intervals, the bucket size will be calculated so that about 1000 buckets span 30 days.`,
		List: &plugin.ListConfig{
			Hydrate: listMetric,
		},
		Get: &plugin.GetConfig{
			KeyColumns: plugin.NewEqualsKeyColumnSlice([]string{"name", "labels", "timestamp"}, plugin.Required),
			Hydrate:    getMetric,
		},
		Columns: []*plugin.Column{
			{Name: "name", Type: proto.ColumnType_STRING, Description: "The metric name used to trigger the aliased promql query."},
			{Name: "labels", Type: proto.ColumnType_JSON, Description: "The label matching to pass into the promql query."},
			{Name: "timestamp", Type: proto.ColumnType_TIMESTAMP, Description: "The timestamp of the calculated metric value."},
			{Name: "value", Type: proto.ColumnType_DOUBLE, Description: "The value of the metric."},
		},
	}
}

func listMetric(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	return nil, fmt.Errorf("unimplemented")
}

func getMetric(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	return nil, fmt.Errorf("unimplemented")
}

func tablePromqlAlias() *plugin.Table {
	return &plugin.Table{
		Name:        "promql_alias",
		Description: "Alias objects are templates for promql requires that are projected as metric names.",
		List: &plugin.ListConfig{
			Hydrate: listAlias,
		},
		Get: &plugin.GetConfig{
			KeyColumns: plugin.SingleColumn("name"),
			Hydrate:    getAlias,
		},
		Columns: []*plugin.Column{
			{Name: "name", Type: proto.ColumnType_STRING, Description: "The name of the query alias."},
			{Name: "template", Type: proto.ColumnType_STRING, Description: "A golang text/template of a promql query."},
		},
	}
}

func listAlias(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	return nil, fmt.Errorf("unimplemented")
}

func getAlias(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	return nil, fmt.Errorf("unimplemented")
}
