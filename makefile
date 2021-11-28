
"promql.plugin": promql/*.go *.go
	go build -o promql.plugin
clean:
	rm -rf steampipe-plugin-promql
	rm -rf promql.plugin

install: promql.plugin
	mkdir -p "$(HOME)/.steampipe/plugins/hub.steampipe.io/plugins/infoblox/promql@latest"
	cp .plugin "$(HOME)/.steampipe/plugins/hub.steampipe.io/plugins/infoblox/promql@latest/promql.plugin"
	mkdir -p $(HOME)/.steampipe/config/
	cp promql.spc $(HOME)/.steampipe/config/promql.spc

