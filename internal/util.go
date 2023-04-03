package internal

func Ptr[T any](x T) *T {
	return &x
}

type contextKey int

const (
	ContextKeyTraceParent contextKey = iota
)
