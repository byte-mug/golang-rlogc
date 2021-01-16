/*
Copyright (C) 2021 Simon Schmidt

This software is provided 'as-is', without any express or implied
warranty.  In no event will the authors be held liable for any damages
arising from the use of this software.

This software is licensed under the Creative Commons "CC0" public domain dedication.
See LICENSE or <http://creativecommons.org/publicdomain/zero/1.0/> for full details.
*/

/*
A simple, non-synchronized RlogC implementation.
*/
package rlogc

import (
	"math"
)

/*
A function that returns a monotinically increasing 64-bit integer.
*/
type Timer func() int64

type HC struct {
	LogC float64
	Time int64
}

/*
This function increments n logarithmically.

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

*/
func F(n float64) float64{
	return math.Max(0,n) + math.Log(math.Exp(-math.Abs(n))+1)
}

/*
Initializes the counter LogC = log(1) and the Time = now.
Called, when an entry enters the cache.

If the decay factor is 0.9, then:

	decay = log(0.9)
*/
func (hc *HC) Enter(now int64, decay float64) {
	hc.LogC = 0 /* Set count to 1 in log space. */
	hc.Time = now
}

/*
Decays the counter and increments it.

If the decay factor is 0.9, then:

	decay = log(0.9)
*/
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
	hc.LogC = F(hc.LogC)
	//if math.IsNaN(hc.LogC) { panic("hc.LogC is NaN") }
}

func fcompare(a, b float64) float64 {
	switch {
	case	math.IsInf(a, 1) && math.IsInf(b, 1),
		math.IsInf(a, -1) && math.IsInf(b, -1):
		return 0
	default:
		return a-b
	}
}

/*
Compares a and b and returns:

	1  if a>b
	-1 if a<b
	0  otherwise

If the decay factor is 0.9, then:

	decay = log(0.9)
*/
func (a *HC) Compare(b *HC, decay float64) int {
	ldiff := fcompare(a.LogC,b.LogC)
	ldiff -= float64(a.Time-b.Time)*decay
	
	switch {
	case ldiff<0: return -1
	case ldiff>0: return 1
	default: return 0
	}
}

//
