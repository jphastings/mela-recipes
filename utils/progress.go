package utils

type progress struct {
	total   int
	channel chan float32
}

func NewProgress(total int) (progress, <-chan float32) {
	pt := progress{
		total:   total,
		channel: make(chan float32),
	}
	return pt, pt.channel
}

func (pt progress) Update(current int) {
	pt.channel <- float32(current) / float32(pt.total)
}
