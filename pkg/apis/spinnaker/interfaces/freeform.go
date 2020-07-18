package interfaces

type FreeForm map[string]interface{}

func (f *FreeForm) DeepCopy() *FreeForm {
	cp := FreeForm{}
	copyInto(*f, cp)
	return &cp
}

func (f *FreeForm) DeepCopyInto(out *FreeForm) {
	// FreeForm is nested and therefore out is already
	// a copy (`*in = *out`). We don't want to copy on top,
	// we want to replace
	m := make(map[string]interface{})
	*out = m
	copyInto(*f, *out)
}

func copyInto(m, cp map[string]interface{}) map[string]interface{} {
	for k, v := range m {
		vm, ok := v.(map[string]interface{})
		if ok {
			n := make(map[string]interface{})
			cp[k] = copyInto(vm, n)
		} else {
			cp[k] = v
		}
	}
	return cp
}
