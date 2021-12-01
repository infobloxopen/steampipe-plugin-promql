package promql

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"text/template"
	"time"

	v1api "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
	"github.com/turbot/steampipe-plugin-sdk/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/plugin"
	"github.com/turbot/steampipe-plugin-sdk/plugin/transform"
)

func tablePromqlMetric() *plugin.Table {

	keyColumns := []*plugin.KeyColumn{
		&plugin.KeyColumn{Name: "name", Operators: []string{"="}, Require: plugin.Required},
		&plugin.KeyColumn{Name: "timestamp", Operators: []string{">", "<", ">=", "<="}, Require: plugin.Optional},
		// &plugin.KeyColumn{Name: "labels", Operators: []string{"@>"}, Require: plugin.Required},
	}
	return &plugin.Table{
		Name:        "promql_metric",
		Description: `Metric executes promql queries defined in the alias table. The queries are sumarized so no more than about 1000 data points come back for each query. For example, if you query for 30 days and the underlying data is sampled at 15 second intervals, the bucket size will be calculated so that about 1000 buckets span 30 days.`,
		List: &plugin.ListConfig{
			Hydrate:    listMetric,
			KeyColumns: keyColumns,
		},
		Get: &plugin.GetConfig{
			KeyColumns: plugin.NewEqualsKeyColumnSlice([]string{"name", "timestamp"}, plugin.Required),
			Hydrate:    getMetric,
		},
		Columns: []*plugin.Column{
			{Name: "name", Type: proto.ColumnType_STRING, Transform: transform.FromField("Metric").TransformP(getMetricLabelFromMetric, "__name__"), Description: "The metric name used to trigger the aliased promql query."},
			{Name: "labels", Type: proto.ColumnType_JSON, Transform: transform.FromField("Metric").TransformP(getLabelsWithoutName, nil), Description: "Map of all labels in the metric."},
			{Name: "timestamp", Type: proto.ColumnType_TIMESTAMP, Transform: transform.FromField("Timestamp").Transform(transform.UnixMsToTimestamp), Description: "Timestamp of the value."},
			{Name: "value", Type: proto.ColumnType_DOUBLE, Description: "Value of the metric."},
		},
	}
}

func getAliasesFromQuery(ctx context.Context, d *plugin.QueryData) ([]string, error) {

	// try to extract names first. These appear as a list of strings
	// when the 'in' operator is used in postgres
	var names []string
	equalQuals := d.KeyColumnQuals
	if list := equalQuals["name"].GetListValue(); list != nil {
		for _, v := range list.GetValues() {
			names = append(names, v.GetStringValue())
		}
	}
	if len(names) > 0 {
		return names, nil
	}
	// no list was found, look for a mandatory = qual
	name := equalQuals["name"].GetStringValue()
	if name == "" {
		return nil, fmt.Errorf("name parameter required")
	}
	return []string{name}, nil
}

func listMetric(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {

	log := plugin.Logger(ctx)

	// metric name(s)
	aliases, err := getAliasesFromQuery(ctx, d)
	if err != nil {
		return nil, err
	}

	// date range
	from, to, err := getTimerange(ctx, d)
	if err != nil {
		return nil, err
	}
	step := (to.Sub(from)/1000 + time.Second/2).Round(time.Second)
	r := v1api.Range{
		Start: from,
		End:   to,
		Step:  step,
	}

	client, err := getPromClient(ctx)
	if err != nil {
		return nil, err
	}

	for _, alias := range aliases {

		queryTemplate, found := getAliasDef(alias)
		if !found {
			queryTemplate = alias + "{ {{.filter}} }"
		}
		tmpl, err := template.New("t").Parse(queryTemplate)
		if err != nil {
			return nil, err
		}

		var buf bytes.Buffer
		if err := tmpl.Execute(&buf, nil); err != nil {
			return nil, fmt.Errorf("could not render query: %s", err)
		}

		result, warnings, err := client.QueryRange(ctx, buf.String(), r)
		if err != nil {
			log.Trace("query", buf.String())
			return nil, fmt.Errorf("could not execute query: %s", err)
		}
		for _, w := range warnings {
			log.Trace("promql.alias", "query_warning", w)
		}
		var rows int
		for _, i := range result.(model.Matrix) {
			i.Metric["__name__"] = model.LabelValue(alias)
			for _, v := range i.Values {
				row := model.Sample{
					Metric:    i.Metric,
					Timestamp: v.Timestamp,
					Value:     v.Value,
				}
				d.StreamListItem(ctx, row)
				rows += 1
			}
		}
		log.Trace("got rows", rows)
	}

	return nil, nil
}

func getTimerange(ctx context.Context, d *plugin.QueryData) (from, to time.Time, err error) {
	quals := d.Quals
	if quals["timestamp"] != nil {
		for _, q := range quals["timestamp"].Quals {
			switch q.Operator {
			case ">":
				from = q.Value.GetTimestampValue().AsTime()
			case ">=":
				from = q.Value.GetTimestampValue().AsTime()
			case "<":
				to = q.Value.GetTimestampValue().AsTime()
			case "<=":
				to = q.Value.GetTimestampValue().AsTime()
			}
		}
	}
	// default to 1 hour if no range is specified
	if from.IsZero() && to.IsZero() {
		to = time.Now()
		from = time.Now().Add(-time.Hour)
	}
	if from.IsZero() {
		return from, to, fmt.Errorf("starting timestamp must be specified: %s", from)
	}
	if to.IsZero() {
		return from, to, fmt.Errorf("ending timestamp must be specified: %s", to)
	}
	return from, to, nil
}

func getMetric(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	return nil, fmt.Errorf("unimplemented get metric")
}

func getLabelsWithoutName(_ context.Context, d *transform.TransformData) (interface{}, error) {
	ls := d.Value.(model.Metric)
	m := make(map[string]string)
	for k, v := range ls {
		if k == "__name__" {
			continue
		}
		m[string(k)] = string(v)
	}
	return m, nil
}

func getMetricLabelFromMetric(_ context.Context, d *transform.TransformData) (interface{}, error) {
	param := d.Param.(string)
	ls := d.Value.(model.Metric)
	return ls[model.LabelName(param)], nil
}

func getLabelFilter(ctx context.Context, d *plugin.QueryData) (filter string, err error) {
	quals := d.Quals
	if quals["labels"] != nil {
		for _, q := range quals["labels"].Quals {
			if q.Operator != "@>" {
				return "", fmt.Errorf("only @> operator supported on labels")
			}
			var m map[string]string
			s := q.Value.GetJsonbValue()
			if err := json.Unmarshal([]byte(s), &m); err != nil {
				return "", fmt.Errorf("could not unmarshal labels condition %s", s)
			}
		}
	}
	if len(filter) > 0 {
		return "", fmt.Errorf("wtf: %s", filter)
	}
	return filter, nil
}
