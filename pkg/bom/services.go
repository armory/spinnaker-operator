package bom

type Service struct {
	Name string
	Java bool
}

var (
	Services = map[string]Service{}
)

func init() {
	// Add oss micro services
	Add(Service{Name: "deck", Java: false})
	Add(Service{Name: "gate", Java: true})
	Add(Service{Name: "orca", Java: true})
	Add(Service{Name: "clouddriver", Java: true})
	Add(Service{Name: "front50", Java: true})
	Add(Service{Name: "rosco", Java: true})
	Add(Service{Name: "igor", Java: true})
	Add(Service{Name: "echo", Java: true})
	Add(Service{Name: "fiat", Java: true})
	Add(Service{Name: "kayenta", Java: true})
}

func Add(service Service) {
	Services[service.Name] = service
}

func JavaServices() []string {
	services := make([]string, 0)
	for name, service := range Services {
		if service.Java {
			services = append(services, name)
		}
	}
	return services
}
