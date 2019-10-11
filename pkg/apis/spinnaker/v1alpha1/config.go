package v1alpha1

func (e *ExposeConfig) GetAggregatedAnnotations(serviceName string) map[string]string {
	annotations := map[string]string{}
	for k, v := range e.Service.Annotations {
		annotations[k] = v
	}
	if c, ok := e.Service.Overrides[serviceName]; ok {
		for k, v := range c.Annotations {
			annotations[k] = v
		}
	}
	return annotations
}
