/*
Copyright (C) 2021 Simon Schmidt

This software is provided 'as-is', without any express or implied
warranty.  In no event will the authors be held liable for any damages
arising from the use of this software.

This software is licensed under the Creative Commons CC0 license.
*/


package leveldb_rlogc

import (
	"github.com/byte-mug/golang-rlogc/rlogc"
	"github.com/syndtr/goleveldb/leveldb/cache"
	"sync"
	"unsafe"
)

var banned = new(rlogc.Element)

type rlcCache struct {
	m sync.Mutex
	heap rlogc.RlogcHeap
	
	capa  int
	space int
}
func (cc *rlcCache) Capacity() int {
	cc.m.Lock(); defer cc.m.Unlock()
	return cc.capa
}
func (cc *rlcCache) SetCapacity(capacity int) {
	cc.m.Lock(); defer cc.m.Unlock()
	cc.capa = capacity
}
func (cc *rlcCache) evict() {
	for cc.space > cc.capa {
		e := cc.heap.EvictNode()
		if e==nil { return } // Queue empty.
		n := (*cache.Node)(e.UserData[0])
		h := (*cache.Handle)(e.UserData[1])
		e.UserData[0] = nil
		e.UserData[1] = nil
		if n!=nil {
			cc.space -= n.Size()
			n.CacheData = nil
		}
		if h!=nil { h.Release() }
	}
}
func (cc *rlcCache) Promote(n *cache.Node) {
	cc.m.Lock(); defer cc.m.Unlock()
	e := (*rlogc.Element)(n.CacheData)
	if e==banned { return }
	if e==nil {
		e = new(rlogc.Element)
		n.CacheData = unsafe.Pointer(e)
		e.UserData[0] = unsafe.Pointer(n)
		e.UserData[1] = unsafe.Pointer(n.GetHandle())
		cc.heap.Enter(e)
		cc.space += n.Size()
		cc.evict()
	} else {
		e.Promote()
	}
	
}
func (cc *rlcCache) nevict(n *cache.Node, ncd unsafe.Pointer) {
	cc.m.Lock(); defer cc.m.Unlock()
	e := (*rlogc.Element)(n.CacheData)
	if e!=banned && e!=nil {
		h := (*cache.Handle)(e.UserData[1])
		e.UserData[0] = nil
		e.UserData[1] = nil
		cc.space -= n.Size()
		if h!=nil { h.Release() }
	}
	n.CacheData = ncd
}
func (cc *rlcCache) Ban(n *cache.Node) {
	cc.nevict(n,unsafe.Pointer(banned))
}
func (cc *rlcCache) Evict(n *cache.Node) {
	cc.nevict(n,nil)
}
func (cc *rlcCache) EvictNS(ns uint64) {
	cc.m.Lock(); defer cc.m.Unlock()
	for _,e := range cc.heap.BorrowElements() {
		n := (*cache.Node)(e.UserData[0])
		h := (*cache.Handle)(e.UserData[1])
		if n==nil { continue }
		if n.NS()!=ns { continue }
		e.UserData[0] = nil
		e.UserData[1] = nil
		{
			cc.space -= n.Size()
			n.CacheData = nil
		}
		if h!=nil { h.Release() }
		/*
		We don't evict the Element, as the .evict()-routine knows
		to handle empty Elements!
		*/
	}
}
func (cc *rlcCache) EvictAll() {
	cc.m.Lock(); defer cc.m.Unlock()
	for _,e := range cc.heap.BorrowElements() {
		n := (*cache.Node)(e.UserData[0])
		h := (*cache.Handle)(e.UserData[1])
		e.UserData[0] = nil
		e.UserData[1] = nil
		if n!=nil {
			cc.space -= n.Size()
			n.CacheData = nil
		}
		if h!=nil { h.Release() }
		/*
		We don't evict the Element, as we already have stolen the whole
		queue from the heap!
		*/
	}
}
func (cc *rlcCache) Close() error {
	return nil
}
var _ cache.Cacher = (*rlcCache)(nil)

func NewRlogC(decay float64, timer rlogc.Timer, initCapacity int) cache.Cacher {
	cc := new(rlcCache)
	cc.heap.Decay = decay
	cc.heap.Timer = timer
	cc.SetCapacity(initCapacity)
	return cc
}

