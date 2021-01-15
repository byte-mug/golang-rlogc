

/*
A simple, non-synchronized RlogC implementation.
*/
package rlogc

import (
	"unsafe"
	"container/heap"
)

/* Cache-Element. */
type Element struct {
	parent *RlogcHeap
	UserData [2]unsafe.Pointer
	
	rank  HC
	index int
}

type elemHeap []*Element
func (e elemHeap) Len() int { return len(e) }
func (e elemHeap) Less(i, j int) bool {
	ei := e[i]
	ej := e[j]
	return ei.rank.Compare(&ej.rank,ei.parent.Decay)<0
}
func (e elemHeap) Swap(i, j int) {
	ei := e[i]
	ej := e[j]
	
	ej.index = i
	ei.index = j
	
	e[i] = ej
	e[j] = ei
}
func (pe *elemHeap) Push(x interface{}) {
	e := *pe
	if cap(e)==len(e) {
		ne := make(elemHeap,len(e),cap(e)*2)
		copy(ne,e)
		e = ne
	}
	el := x.(*Element)
	el.index = len(e)
	*pe = append(e,el)
}
func (pe *elemHeap) Pop() interface{} {
	e := *pe
	i := len(e)-1
	*pe = e[:i]
	el := e[i]
	el.index = -1
	e[i] = nil
	return el
}

type RlogcHeap struct {
	Decay float64
	Timer Timer
	
	queue elemHeap
}
/*
Enters a node to the heap.
*/
func (r *RlogcHeap) Enter(e *Element) {
	if e.parent!=nil { return }// Error!
	e.parent = r
	e.rank.Enter(r.Timer(),r.Decay)
	e.index = r.queue.Len()
	heap.Push(&r.queue,e)
}

/*
Evicts the least recently/frequently used element and returns it.
*/
func (r *RlogcHeap) EvictNode() (e *Element) {
	if r.queue.Len()==0 { return nil }
	e = heap.Pop(&r.queue).(*Element)
	e.parent = nil
	return
}

// XXX: DEBUG
//func (r *RlogcHeap) Len() int { return r.queue.Len() }

/*
Evicts the node from the RlogcHeap containing it.
*/
func (e *Element) Evict() {
	if e.parent==nil { return }
	heap.Remove(&e.parent.queue,e.index)
	e.parent = nil
}

/*
A cache-hit on this element.
*/
func (e *Element) Promote() {
	if e.parent==nil { return }
	r := e.parent
	e.rank.Access(r.Timer(),r.Decay)
	heap.Fix(&r.queue,e.index)
}


/*
Borrows the priority queue. DANGEROUS! Handle with care!
*/
func (r *RlogcHeap) BorrowElements() []*Element {
	return r.queue
}
/*
Borrows the priority queue. DANGEROUS! Handle with care!
*/
func (r *RlogcHeap) StealElements() []*Element {
	elems := r.queue
	r.queue = r.queue[:0]
	return elems
}

