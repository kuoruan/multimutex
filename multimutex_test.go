package multimutex

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestMultiMutex(t *testing.T) {
	wg := sync.WaitGroup{}

	const number = 2

	var ml MultiMutex

	for i := 0; i < number*5; i++ {
		wg.Add(1)

		go func(ii int) {
			key := ii % number

			now := time.Now().Unix()
			fmt.Printf("[%d] %d - %d\n", key, ii, now)

			ml.Lock(key)
			fmt.Printf("->  [%d] %d - %d\n", key, ii, now)
			ml.Unlock(key)

			wg.Done()
		}(i)
	}
	wg.Wait()
}
