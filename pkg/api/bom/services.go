package bom

// Struct for describing the spinnaker micro services that
// the operator might install.
// Useful for doing transformations o generating configs
// based on known services.
type Service struct {
	Name string
	Type string
}

var (
	Services = map[string]Service{}
)

func init() {
	// Add oss micro services
	Add(Service{Name: "deck", Type: "ui"})
	Add(Service{Name: "gate", Type: "java"})
	Add(Service{Name: "orca", Type: "java"})
	Add(Service{Name: "clouddriver", Type: "java"})
	Add(Service{Name: "front50", Type: "java"})
	Add(Service{Name: "rosco", Type: "java"})
	Add(Service{Name: "igor", Type: "java"})
	Add(Service{Name: "echo", Type: "java"})
	Add(Service{Name: "fiat", Type: "java"})
	Add(Service{Name: "kayenta", Type: "java"})
}

func Add(service Service) {
	Services[service.Name] = service
}

func JavaServices() []string {
	services := make([]string, 0)
	for name, service := range Services {
		if service.Type == "java" {
			services = append(services, name)
		}
	}
	return services
}
