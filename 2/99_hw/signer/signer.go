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
	defer close(chansInput[0])

	var wg sync.WaitGroup
	for i := range jobs {
		wg.Add(1)

		i := i
		go func() {
			defer wg.Done()
			defer close(chansInput[i+1])
			jobs[i](chansInput[i], chansInput[i+1])
		}()
	}

	wg.Wait()

	return nil
}

var SingleHash job = func(in, out chan interface{}) {
	md5Mutex := sync.Mutex{}
	wg := sync.WaitGroup{}
	for input := range in {
		wg.Add(1)
		input := input
		go func() {
			defer wg.Done()

			outChanCrc32 := make(chan string, 0)
			go CalcDataSignerCrc32(fmt.Sprint(input), outChanCrc32)

			md5Mutex.Lock()
			md5Input := DataSignerMd5(fmt.Sprint(input))
			md5Mutex.Unlock()

			outChanMd5 := make(chan string, 0)
			go CalcDataSignerCrc32(md5Input, outChanMd5)

			inSigner := <-outChanCrc32
			md5Signer := <-outChanMd5

			out <- inSigner + "~" + md5Signer
		}()
	}

	wg.Wait()
}

var MultiHash job = func(in, out chan interface{}) {
	var mainWg sync.WaitGroup
	for input := range in {
		mainWg.Add(1)
		input := input
		go func() {
			defer mainWg.Done()
			var wg sync.WaitGroup
			var results = make([]string, 6)
			for i := 0; i <= 5; i++ {
				wg.Add(1)
				go func(i int) {
					defer wg.Done()
					crc32Chan := make(chan string, 0)
					go CalcDataSignerCrc32(fmt.Sprint(i)+input.(string), crc32Chan)
					results[i] = <-crc32Chan
				}(i)
			}

			wg.Wait()
			out <- strings.Join(results, "")
		}()
	}
	mainWg.Wait()
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
