package callcounter

type CallCounter interface {
	Call(name string)
	Count(name string) int64
}
