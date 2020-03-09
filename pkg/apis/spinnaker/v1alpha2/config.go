package v1alpha2

func (e *ExposeConfig) GetAggregatedAnnotations(serviceName string) map[string]string {
	annotations := map[string]string{}
	for k, v := range e.GetService().GetAnnotations() {
		annotations[k] = v
	}
	if c, ok := e.GetService().GetOverrides()[serviceName]; ok {
		for k, v := range c.GetAnnotations() {
			annotations[k] = v
		}
	}
	return annotations
}
