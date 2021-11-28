package main

import (
    "github.com/turbot/steampipe-plugin-sdk/plugin"
    "github.com/infobloxopen/steampipe-plugin-promql/promql"
)

func main() {
    plugin.Serve(&plugin.ServeOpts{PluginFunc: promql.Plugin})
}

