# The *R+log(C)* Cache policy.


Design and implementation of the *R+log(C)* Cache policy as published 15.01.2021, 22:06.

The *R+log(C)* Cache policy can be considered a variant of *practical LFUDA*. The algorithm maintains logarithmic access-counter for each item stored in cache, the counter is discarded once the item is evicted. Each item is monotonically aged.

## The Ranking.

A ranking value is defined as:

```go
type HC struct {
	LogC float64 // Logarithmic Hit-Counter
	Time int64   // Timestamp
}
```

On each page hit the following function is called. `decay` is the amount of decay per time unit. `now` is the current timestamp.

```go
func (hc *HC) Access(now int64, decay float64) {
	/*
	Apply the difference between the timestamp and NOW
	to LogC and update the timestamp to NOW.
	*/
	hc.LogC += float64(now-hc.Time)*decay
	hc.Time = now
	
	/*
	Logarithmically increment LogC.
	*/
	hc.LogC = F(hc.LogC) // use a function called f(n) = ...
}
```

Upon initializing the ranking value for a new item, do:

```go
func (hc *HC) Enter(now int64, decay float64) {
	hc.LogC = 0 /* Set count to 1 in log space. */
	hc.Time = now
}
```

Comparison is defined as:

```go
func (a *HC) Compare(b *HC, decay float64) int {
	anchor := a.Time // or any other anchor.
	rank_a :=  a.LogC + (decay*(a.Time-anchor))
	rank_b :=  b.LogC + (decay*(b.Time-anchor))
    
	switch {
	case rank_a<rank_b: return -1
	case rank_a>rank_b: return 1
	default: return 0
	}
}
```

## The Function *f(n)*


This function computes (in a mathematical sense):

	f(n) = log(exp(n)+1)

However, to prevent floating point overflows, this function, its implementation
exploits the following rule:

	f(n) = n + f(-n)

Which can be proven as follows:

	f(n) = log(exp(n)+1)
	f(n) = log(exp(n)+exp(0))
	f(n) = log(exp(n-n)+exp(0-n))+n
	f(n) = log(exp(0)+exp(-n))+n
	f(n) = log(1+exp(-n))+n
	f(n) = n + log(exp(-n)+1)
	f(n) = n + f(-n)

The function is implemented as

	f(n) = max(0,n) + log(exp(-abs(n))+1)

Which is correct because:

	max(0,n) = 0 IF n<0
	-abs(n)  = n IF n<0

	max(0,n) = n IF n>0
	-abs(n) = -n IF n>0

	max(0,n) = -abs(n) = 0 IF n = 0

## The Priority Queue.

My implementation uses a data structure called a <a href="https://en.wikipedia.org/wiki/Heap_(data_structure)">heap</a>.

