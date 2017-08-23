package health2

type Stat interface {
	Update(value float64)
}
