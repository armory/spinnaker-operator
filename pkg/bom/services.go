package bom

type Service struct {
	Java bool
}

var (
	Services = map[string]Service{}
)

func init() {
	// Add oss micro services
	Add("deck", Service{Java: false})
	Add("gate", Service{Java: true})
	Add("orca", Service{Java: true})
	Add("clouddriver", Service{Java: true})
	Add("front50", Service{Java: true})
	Add("rosco", Service{Java: true})
	Add("igor", Service{Java: true})
	Add("echo", Service{Java: true})
	Add("fiat", Service{Java: true})
	Add("kayenta", Service{Java: true})
}

func Add(serviceName string, service Service) {
	Services[serviceName] = service
}
