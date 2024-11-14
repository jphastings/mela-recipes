package formats

import "sync"

var AvailableFormats []Format
var mu sync.Mutex

func Register(format Format) {
	mu.Lock()
	AvailableFormats = append(AvailableFormats, format)
	mu.Unlock()
}
