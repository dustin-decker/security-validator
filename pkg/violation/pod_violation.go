package violation

type PodViolation struct {
	PodName   string
	Violation string
	Error     error
}
