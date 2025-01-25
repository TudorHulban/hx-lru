package hxlru

func deepCopySlice[V any](source []V) []V {
	result := make([]V, len(source), len(source))

	copy(result, source)

	return result
}
