package webrouter

func hasSameMethod(methods []string, method string) bool {
	for _, mtd := range methods {
		if mtd == method {
			return true
		}
	}

	return false
}
