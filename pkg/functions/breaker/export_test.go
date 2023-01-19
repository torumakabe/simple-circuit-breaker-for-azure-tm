package breaker

var (
	SortEndpoints = sortEndpoints
)

func SetTMBreaker(b Breaker) (resetFunc func()) {
	var tmp Breaker
	tmp, defaultTMBreaker = defaultTMBreaker, b
	return func() {
		defaultTMBreaker = tmp
	}
}
