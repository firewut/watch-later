package helpers

func ExistsMapKey(m map[string]interface{}, k string) bool {
	_, ok := m[k]
	return ok
}

func ExistsMapKeyString(m map[string]string, k string) bool {
	_, ok := m[k]
	return ok
}
