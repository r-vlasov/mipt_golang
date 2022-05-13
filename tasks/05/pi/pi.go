package pi

import (
	"sync"
)

const (
	Rsize = 1.0
)

func CalculatePi(concurrent int, iterations int, gen RandomPointGenerator) float64 {
	var wg sync.WaitGroup
	wg.Add(concurrent)

	// worker communication channel
	workerChannel := make(chan int)

	// worker function
	piWorker := func(workerIterations int) {
		inCircle := 0
		for i := 0; i < workerIterations; i++ {
			x, y := gen.Next()
			if x * x + y * y <= Rsize * Rsize {
				inCircle += 1
			}
		}
		workerChannel <- inCircle

		defer wg.Done()
	}

	// total count of iterations = iterations
	// we should schedule this iterations between concurrent workers, but iteration % concurrent could be != 0 
	for i := 0; i < concurrent; i++ {
		workerIterations := iterations / concurrent
		if iterations % concurrent - i > 0 {
			workerIterations += 1
		}
		go piWorker(workerIterations)
	}

	// accumulate all results
	totalInCircle := 0
	for i := 0; i < concurrent; i++ {
		totalInCircle += <- workerChannel
	}

	// waiting
	wg.Wait()
	close(workerChannel)

	// 4 - is a constant of normalization (in 1x1 quadrate located circle with pi / 4 square)
	return 4 * float64(totalInCircle) / float64(iterations)
}

type RandomPointGenerator interface {
	Next() (float64, float64)
}
