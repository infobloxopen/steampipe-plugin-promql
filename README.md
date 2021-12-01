# steampipe-plugin-promql
steampipe promql plugin

Configuration
-------------

The plugin is configured via a config file located at:
* /etc/steampipe-promql/config.yaml
* /usr/local/etc/steampipe-promql/config.yaml
* $(HOME)/.steampipe-promql/config.yaml

The endpoint key controls what Prometheus endpoint is used by
the plugin.

The alias key is a map of promql queries that are projected
as metrics in the `metrics` table.


Tables
------

promql.promql_alias
===================

The alias table contains two columns: name & value. The value
is a promql query template. Its used by the `metric` table to
execute / project that query as a metric. Aliases are configured
via the config file: ~/.steampipe-promql/config.yaml. 

promql.promql_metric
====================

The metric table contains 4 colums: name, labels, timestamp and value. The
plugin forces the user to provide a name in the query. This name is a reference
to an Alias. That alias is a promql template that's used to execute a query
on a Prometheus compatible server. This allows the user to use a single table
to fetch any queries they want and hide the details about how those metrics
are projected. 

Additionally, step is calculated to be at least 1 second or 1/1000ths of the time
span that is being queried. This allows for the output to be reasonably graphed
without having to calculate the step size explicitly. 
