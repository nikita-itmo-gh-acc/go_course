package main

import (
	"sort"
	"strconv"
	"strings"
	"sync"
)

func ExecutePipeline(workers ...job) {
	channels := make([]chan interface{}, len(workers)+1)
	channels[0] = make(chan interface{}, 100)
	wg := new(sync.WaitGroup)
	wg.Add(len(workers))
	for i, worker := range workers {
		channels[i+1] = make(chan interface{})
		go func(w job, in chan interface{}, out chan interface{}) {
			w(in, out)
			wg.Done()
			close(out)
		}(worker, channels[i], channels[i+1])
	}
	wg.Wait()
}

func SingleHash(inCh chan interface{}, outCh chan interface{}) {
	waitAll := new(sync.WaitGroup)
	for data := range inCh {
		data := data.(int)
		mutex := new(sync.Mutex)
		mutex.Lock()
		md_hash := DataSignerMd5(strconv.Itoa(data))
		mutex.Unlock()
		var crcBase, crcMd string
		crc32 := func(crcRes *string, dataToHash string, w *sync.WaitGroup) {
			defer w.Done()
			*crcRes = DataSignerCrc32(dataToHash)
		}
		waitAll.Add(1)
		go func() {
			wg := new(sync.WaitGroup)
			wg.Add(2)
			go crc32(&crcBase, strconv.Itoa(data), wg)
			go crc32(&crcMd, md_hash, wg)
			wg.Wait()
			outCh <- crcBase + "~" + crcMd
			waitAll.Done()
		}()
	}
	waitAll.Wait()
}

func MultiHash(inCh chan interface{}, outCh chan interface{}) {
	waitAll := new(sync.WaitGroup)
	for data := range inCh {
		data := data.(string)
		crc32 := func(dataToHash string, it int, w *sync.WaitGroup, m *sync.Mutex, syncCh chan int, r *string) {
			crcRes := DataSignerCrc32(dataToHash)
			for signal := range syncCh {
				if signal == it {
					m.Lock()
					*r += crcRes
					m.Unlock()
					syncCh <- it + 1
					w.Done()
					return
				} else {
					syncCh <- signal
				}
			}
		}
		waitAll.Add(1)
		go func() {
			var res string
			crcSync := make(chan int, 1)
			crcSync <- 0
			wg := new(sync.WaitGroup)
			mutex := new(sync.Mutex)
			for th := 0; th < 6; th++ {
				wg.Add(1)
				go crc32(strconv.Itoa(th)+data, th, wg, mutex, crcSync, &res)
			}
			wg.Wait()
			close(crcSync)
			outCh <- res
			waitAll.Done()
		}()
	}
	waitAll.Wait()
}

func CombineResults(inCh chan interface{}, outCh chan interface{}) {
	results := make([]string, 0, 100)
	for data := range inCh {
		results = append(results, data.(string))
	}
	sort.Slice(results, func(i, j int) bool { return results[i] < results[j] })
	outCh <- strings.Join(results, "_")
}
