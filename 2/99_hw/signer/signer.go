package main

import (
	"fmt"
	"sort"
	"strings"
	"sync"
)

func ExecutePipeline(jobs ...job) error {
	chansInput := make([]chan interface{}, len(jobs)+1)
	for i := range chansInput {
		chansInput[i] = make(chan interface{}, 0)
	}

	var wg sync.WaitGroup
	for i := range jobs {
		wg.Add(1)

		i := i
		go func() {
			jobs[i](chansInput[i], chansInput[i+1])
			close(chansInput[i+1])
			wg.Done()
		}()
	}

	wg.Wait()

	return nil
}

var SingleHash job = func(in, out chan interface{}) {
	input := <-in
	output := DataSignerCrc32(input.(string)) + "~" + DataSignerCrc32(DataSignerMd5(input.(string)))
	out <- output
}

var MultiHash job = func(in, out chan interface{}) {
	input := <-in
	var results = make([]string, 5)
	for i := 0; i <= 5; i++ {
		results[i] = DataSignerCrc32(fmt.Sprint(i) + input.(string))
	}
	out <- strings.Join(results, "")
}

var CombineResults job = func(in, out chan interface{}) {
	var results = make([]string, 0)
	for result := range in {
		results = append(results, result.(string))
	}
	sort.Sort(sort.StringSlice(results))
	out <- strings.Join(results, "_")
}
