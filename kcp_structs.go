package spine

type serviceRequest[K any, V any] struct {
	input  K
	output chan serviceOutput[V]
}

type serviceOutput[V any] struct {
	data V
	err  error
}
