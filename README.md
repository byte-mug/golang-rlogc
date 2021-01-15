# golang-rlogc
R+log(C) Cache policy. I am the author of this algorithm BTW.

## Why another cache algorithm?

I was annoyed of the lack of proper cache algorithms, so I invented my own.

Whereas other cache policies like *LRU-2*, *2Q*, *LIRS* and *ARC* are implemented using two
queues and/or a victim cache. *R+log(C)*, on the other hand, only requires a single priority queue.

#### May I rip off this cache policy?

Please do!

#### And what about commercial applications?

Great! I'd appreciate it.
