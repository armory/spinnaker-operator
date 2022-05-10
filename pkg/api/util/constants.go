package util

const (
	GateServiceName         = "spin-gate"
	GateX509ServiceName     = "spin-gate-x509"
	GateX509PortName        = "gate-x509"
	GateDefaultPort         = int32(8084)
	GateOverrideBaseUrlProp = "security.apiSecurity.overrideBaseUrl"
	GateSSLEnabledProp      = "security.apiSecurity.ssl.enabled"
	DeckServiceName         = "spin-deck"
	DeckOverrideBaseUrlProp = "security.uiSecurity.overrideBaseUrl"
	DeckSSLEnabledProp      = "security.uiSecurity.ssl.enabled"
	DeckDefaultPort         = int32(9000)
	ClouddriverName         = "clouddriver"
)
