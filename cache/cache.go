/*
Copyright (C) 2021 Simon Schmidt

This software is provided 'as-is', without any express or implied
warranty.  In no event will the authors be held liable for any damages
arising from the use of this software.

This software is licensed under the Creative Commons "CC0" public domain dedication.
See LICENSE or <http://creativecommons.org/publicdomain/zero/1.0/> for full details.
*/

/*
This package implements a fixed-size thread safe *R+log(C)* cache.
*/
package cache

import (
	"github.com/byte-mug/golang-rlogc/rlogc"
	"sync"
	"unsafe"
)

/*
EvictCB is used to get a callback when a cache entry is evicted.
*/
type EvictCB func(key , value interface{})

func nullcb(key , value interface{}) {}

type entry struct{
	e rlogc.Element
	key, value interface{}
}

type Cache struct{
	m     sync.Mutex
	heap  rlogc.RlogcHeap
	evcb  EvictCB
	size  int
	
	index map[interface{}]*entry
}

func NewCache(decay float64, timer rlogc.Timer, size int, cb EvictCB) *Cache {
	if cb==nil { cb = nullcb }
	c := new(Cache)
	
	c.heap.Decay = decay
	c.heap.Timer = timer
	c.evcb = cb
	c.size = size
	
	return c
}

/*
Adds or replaces a cache entry.
*/
func (c *Cache) Add(key, value interface{}) (evicted bool) {
	c.m.Lock(); defer c.m.Unlock()
	
	e,ok := c.index[key]
	if ok {
		e.key,e.value = key,value
		e.e.Promote()
		return
	}
	
	e = new(entry)
	e.key,e.value = key,value
	e.e.UserData[0] = unsafe.Pointer(e)
	c.heap.Enter(&e.e)
	
	c.index[key] = e
	
	if c.heap.Len() <= c.size { return }
	
	el := c.heap.EvictNode()
	if el==nil { return } // safety first.
	e = (*entry)(el.UserData[0])
	
	evicted = true
	
	delete(c.index,e.key)
	c.evcb(e.key,e.value)
	
	return
}

/*
Looks up a cache entry.
*/
func (c *Cache) Get(key interface{}) (value interface{}, ok bool) {
	c.m.Lock(); defer c.m.Unlock()
	
	var e *entry
	e,ok = c.index[key]
	if ok {
		value = e.value
		e.e.Promote()
	}
	return
}

/*
Looks up a cache entry without updating the cache statistics.
*/
func (c *Cache) Peek(key interface{}) (value interface{}, ok bool) {
	c.m.Lock(); defer c.m.Unlock()
	
	var e *entry
	e,ok = c.index[key]
	if ok { value = e.value }
	return
}

/*
Returns the number of entries.
*/
func (c *Cache) Len() int {
	c.m.Lock(); defer c.m.Unlock()
	return c.heap.Len()
}

/*
Removes a specific cache entry.
*/
func (c *Cache) Remove(key interface{}) (present bool) {
	c.m.Lock(); defer c.m.Unlock()
	
	e,ok := c.index[key]
	if ok {
		present = true
		e.e.Evict()
		c.evcb(e.key,e.value)
		delete(c.index,key)
	}
	return
}

/*
Looks up a cache entry without updating the cache statistics and witout returning it's value.
*/
func (c *Cache) Contains(key interface{}) bool {
	c.m.Lock(); defer c.m.Unlock()
	
	_,ok := c.index[key]
	return ok
}

/*
Looks up a cache entry without updating the cache statistics and adds a new entry, if no cache entry was found.
*/
func (c *Cache) PeekOrAdd(key, value interface{}) (previous interface{}, ok, evicted bool) {
	c.m.Lock(); defer c.m.Unlock()
	
	var e *entry
	e,ok = c.index[key]
	if ok { return e.value,true,false }
	
	e = new(entry)
	e.key,e.value = key,value
	e.e.UserData[0] = unsafe.Pointer(e)
	c.heap.Enter(&e.e)
	
	c.index[key] = e
	
	if c.heap.Len() <= c.size { return }
	
	el := c.heap.EvictNode()
	if el==nil { return } // safety first.
	e = (*entry)(el.UserData[0])
	
	evicted = true
	
	delete(c.index,e.key)
	c.evcb(e.key,e.value)
	
	return
}

func (c *Cache) GetOldest() (key interface{}, value interface{}, ok bool) {
	c.m.Lock(); defer c.m.Unlock()
	
	es := c.heap.BorrowElements()
	if len(es)==0 { return }
	e := (*entry)(es[0].UserData[0])
	key,value,ok = e.key,e.value,true
	return
}
func (c *Cache) RemoveOldest() (key interface{}, value interface{}, ok bool) {
	c.m.Lock(); defer c.m.Unlock()
	
	el := c.heap.EvictNode()
	if el==nil { return }
	e := (*entry)(el.UserData[0])
	
	key,value,ok = e.key,e.value,true
	delete(c.index,key)
	c.evcb(key,value)
	
	return
}

func (c *Cache) ContainsOrAdd(key, value interface{}) (ok, evicted bool) {
	_,ok,evicted = c.PeekOrAdd(key,value)
	return
}
func (c *Cache) Resize(size int) (evicted int) {
	c.size = size
	
	for c.size < c.heap.Len() {
		el := c.heap.EvictNode()
		if el==nil { return }
		e := (*entry)(el.UserData[0])
		
		evicted ++
		delete(c.index,e.key)
		c.evcb(e.key,e.value)
	}
	return
}
