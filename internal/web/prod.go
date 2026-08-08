package web

import "strings"

// prodHostMarker identifies the Railway production pooler in a DSN. The
// admin panel's destructive actions (sync apply, migrations) are hard-blocked
// against prod unless WEB_ALLOW_PROD_MUTATIONS=true — see Config.
const prodHostMarker = "roundhouse.proxy.rlwy.net"

// IsProdDSN reports whether a database DSN points at the production pooler.
func IsProdDSN(dsn string) bool {
	return strings.Contains(dsn, prodHostMarker)
}
