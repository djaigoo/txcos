// Package utils utils

package utils

import "sync"

type TokenBucket struct {
    ch chan struct{}
    wg *sync.WaitGroup
}

func NewTokenBucket(size int) *TokenBucket {
    tb := &TokenBucket{}
    tb.ch = make(chan struct{}, size)
    for i := 0; i < size; i++ {
        tb.ch <- struct{}{}
    }
    tb.wg = new(sync.WaitGroup)
    return tb
}

func (tb *TokenBucket) Get() {
    tb.wg.Add(1)
    <-tb.ch
}

func (tb *TokenBucket) Put() {
    tb.ch <- struct{}{}
    tb.wg.Done()
}

func (tb *TokenBucket) Wait() {
    tb.wg.Wait()
}
