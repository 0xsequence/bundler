package utils

import "sync"

func FanIn[T any](channels ...<-chan T) <-chan T {
	var wg sync.WaitGroup
	out := make(chan T)

	send := func(c <-chan T) {
		for msg := range c {
			out <- msg
		}
		wg.Done()
	}

	wg.Add(len(channels))
	for _, c := range channels {
		go send(c)
	}

	go func() {
		wg.Wait()
		close(out)
	}()

	return out
}
