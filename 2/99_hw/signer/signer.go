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
		chansInput[i] = make(chan interface{})
	}

	var wg sync.WaitGroup
	for i := range jobs {
		wg.Add(1)

		i := i
		go func() {
			defer wg.Done()
			jobs[i](chansInput[i], chansInput[i+1])
			close(chansInput[i+1])
		}()
	}

	wg.Wait()

	return nil
}

var SingleHash job = func(in, out chan interface{}) {
	md5Mutex := sync.Mutex{}
	for input := range in {
		md5Mutex.Lock()
		md5Signer := DataSignerMd5(fmt.Sprint(input))
		md5Mutex.Unlock()

		outChan := make(chan string, 0)
		go CalcDataSignerCrc32(md5Signer, outChan)
		md5Signer = <-outChan

		go CalcDataSignerCrc32(fmt.Sprint(input), outChan)
		inSigner := <-outChan

		out <- inSigner + "~" + md5Signer
	}
}

var MultiHash job = func(in, out chan interface{}) {
	for input := range in {
		var wg sync.WaitGroup
		var results = make([]string, 6)
		for i := 0; i <= 5; i++ {
			i := i
			wg.Add(1)
			go func() {
				crc32Chan := make(chan string, 0)
				go CalcDataSignerCrc32(fmt.Sprint(i)+input.(string), crc32Chan)
				results[i] = <-crc32Chan
				wg.Done()
			}()
		}

		wg.Wait()
		out <- strings.Join(results, "")
	}
}

var CombineResults job = func(in, out chan interface{}) {
	var results = make([]string, 0)
	for result := range in {
		results = append(results, result.(string))
	}
	sort.Sort(sort.StringSlice(results))
	out <- strings.Join(results, "_")
}

func CalcDataSignerCrc32(input string, out chan string) {
	out <- DataSignerCrc32(input)
}
