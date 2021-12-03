package promql

import (
	"context"
	"fmt"

	"github.com/turbot/steampipe-plugin-sdk/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/plugin"
)

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

type Alias struct {
	Name     string `json:"name"`
	Template string `json:"template"`
}

func listAlias(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	cfg, err := getConfig()
	if err != nil {
		return nil, err
	}
	for k, v := range cfg.Aliases {
		d.StreamListItem(ctx, Alias{Name: k, Template: v})
	}
	return nil, nil
}

func getAlias(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	cfg, err := getConfig()
	if err != nil {
		return nil, err
	}
	quals := d.KeyColumnQuals
	plugin.Logger(ctx).Trace("getAlias", "quals", quals)
	name := quals["name"].GetStringValue()
	plugin.Logger(ctx).Trace("getAlias", "name", name)
	v, ok := cfg.Aliases[name]
	if !ok {
		return nil, fmt.Errorf("could not find item")
	}
	return &Alias{Name: name, Template: v}, nil
}
