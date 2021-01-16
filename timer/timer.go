/*
Copyright (C) 2021 Simon Schmidt

This software is provided 'as-is', without any express or implied
warranty.  In no event will the authors be held liable for any damages
arising from the use of this software.

This software is licensed under the Creative Commons "CC0" public domain dedication.
See LICENSE or <http://creativecommons.org/publicdomain/zero/1.0/> for full details.
*/

/* Timer implementations for RlogC */
package timer

import "time"
import "github.com/byte-mug/golang-rlogc/rlogc"

/*
A monotonic timer, that increases once every second.

Calls to this function are somewhat expensive, because this function calls

	time.Now().Unix()

*/
func Seconds() rlogc.Timer {
	off := time.Now().Unix()
	return func() int64 {
		return time.Now().Unix()-off
	}
}


type interval struct {
	ticker <-chan time.Time
	value  int64
}
func (i *interval) run() {
	for {
		<- i.ticker
		i.value ++
	}
}
func (i *interval) now() int64 {
	return i.value
}

/*
A monotonic timer that increases in a given interval.

Calls to this function are cheap, because it only fetches a field of a struct
that is incremented by a goroutine, that runs in background.

WARNING: this function creates a goroutine, that won't stop!
Do only call this a bounded number of times!
*/
func Interval(d time.Duration) rlogc.Timer {
	i := new(interval)
	i.ticker = time.Tick(d)
	go i.run()
	return i.now
}

type incrementer struct {
	value int64
}
func (i *incrementer) now() int64 {
	i.value++
	return i.value
}

/*
A pseudo-timer, that increases every time it is called.
*/
func Increment() rlogc.Timer {
	i := new(incrementer)
	return i.now
}


