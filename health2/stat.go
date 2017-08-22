package health2

type Stat interface {
	Set(delta float64)
	Add(delta float64)
	Observe(delta float64)
	Increment()
	Decrement()
}
