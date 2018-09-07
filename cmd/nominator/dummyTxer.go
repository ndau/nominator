package main

import (
	"fmt"
	"math"
	"math/rand"
	"time"
)

// dummyTxer implements the Txer type but just fakes listening for tx data and posting
type dummyTxer struct {
	timing time.Duration
	timer  *LoopTimer
}

var _ = ndauTxer(&dummyTxer{})

// Post implements Txer for dummyTxer
func (n *dummyTxer) Post(r int64) error {
	fmt.Printf("POSTING tx with rand %d at %s\n", r, time.Now())
	return nil
}

// Listen implements Txer for dummyTxer.
func (n *dummyTxer) Listen(nomChan chan Nomination) {
	n.timer = NewLoopTimer(time.Second, n.timing, func() {
		nomChan <- Nomination{timestamp: time.Now(), r: rand.Int63n(math.MaxInt64)}
	})
}
