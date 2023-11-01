package vector

type Vector struct {
	ID   string
	Data []float32
}

func (v *Vector) Dim() int {
	return len(v.Data)
}
