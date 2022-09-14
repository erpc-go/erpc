package center

type funcs map[string]bool

func (f funcs) addFunc(funcs []string) {
	for _, v := range funcs {
		if !f[v] {
			f[v] = true
		}
	}
}

func (f funcs) empty() bool {
	return len(f) == 0
}
