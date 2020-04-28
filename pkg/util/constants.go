package util

const (
	GateServiceName         = "spin-gate"
	GateX509ServiceName     = "spin-gate-x509"
	GateX509PortName        = "gate-x509"
	GateOverrideBaseUrlProp = "security.apiSecurity.overrideBaseUrl"
	GateSSLEnabledProp      = "security.apiSecurity.ssl.enabled"
	DeckServiceName         = "spin-deck"
	DeckOverrideBaseUrlProp = "security.uiSecurity.overrideBaseUrl"
	DeckSSLEnabledProp      = "security.uiSecurity.ssl.enabled"
	ClouddriverName         = "clouddriver"
)

var (
	SpinnakerServices = map[string]struct{}{
		"clouddriver": {},
		"orca":        {},
		"echo":        {},
		"fiat":        {},
		"igor":        {},
		"rosco":       {},
		"front50":     {},
		"kayenta":     {},
		"gate":        {},
		"dinghy":      {},
		"terraformer": {},
	}
	SpinnakerJavaServices = map[string]struct{}{
		"clouddriver": {},
		"orca":        {},
		"echo":        {},
		"fiat":        {},
		"igor":        {},
		"rosco":       {},
		"front50":     {},
		"kayenta":     {},
		"gate":        {},
	}
	SpinnakerGoServices = map[string]struct{}{
		"dinghy":      {},
		"terraformer": {},
	}
)
