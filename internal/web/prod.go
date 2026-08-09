package web

import "strings"

// prodHostMarker identifies the Railway production pooler in a DSN. The
// admin panel's production banner and operational context.
const prodHostMarker = "roundhouse.proxy.rlwy.net"

// IsProdDSN reports whether a database DSN points at the production pooler.
// The substring check is intentionally conservative: ENVIRONMENT=production
// is ALSO treated as prod (see IsProd) so a renamed pooler host cannot
// silently disable the guard.
func IsProdDSN(dsn string) bool {
	return strings.Contains(dsn, prodHostMarker)
}

// IsProd reports whether the panel is running against production: either the
// DSN is the known pooler or ENVIRONMENT=production is set.
func IsProd(dsn, environment string) bool {
	return IsProdDSN(dsn) || strings.EqualFold(environment, "production")
}
