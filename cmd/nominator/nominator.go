package main

import (
	cryptorand "crypto/rand"
	"fmt"
	"math"
	"math/big"
	"math/rand"
	"os"
	"os/signal"
	"time"

	arg "github.com/alexflint/go-arg"
	"github.com/ndau/o11y/pkg/honeycomb"
	log "github.com/sirupsen/logrus"
)

type Nomination struct {
	timestamp time.Time
	r         int64
}

type ndauTxer interface {
	Post(r int64) error
	Listen(chan Nomination)
}

// start a timer with the wakeup time being a random value between min and max since the last nomination tx
// watch for new node nominations being posted to the blockchain
//     if it happens, restart the timer
// when timer expires:
//     create a random number in the range (0, maxInt64]
//     post it to the blockchain as part of a new node nomination tx
//     restart the timer (normally, we'll see the node nomination tx come through on the other channel and that will restart it
//     but we don't want to fail to have a timer running in the case that the tx fails for some other reason)
func run(txer ndauTxer, minTime, maxTime time.Duration, logger *log.Entry) error {
	nomChan := make(chan Nomination, 0)

	// handle a ctrl C to shut down gracefully
	signalChan := make(chan os.Signal, 0)
	signal.Notify(signalChan, os.Interrupt)

	// and ensure the graceful shutdown
	defer close(nomChan)
	defer close(signalChan)

	// our cycle time can be non-cryptographically random
	getCycleTime := func() time.Duration {
		return maxTime + time.Duration(rand.Intn(int(maxTime-minTime)))
	}
	timer := time.NewTimer(getCycleTime())

	txer.Listen(nomChan)

	for {
		select {
		// if the timer runs out we need to post a tx
		case <-timer.C:
			r, err := cryptorand.Int(cryptorand.Reader, big.NewInt(math.MaxInt64))
			if err != nil {
				logger.WithError(err).Error("cryptorand returned error")
			} else {
				txer.Post(r.Int64())
			}
			timer.Reset(getCycleTime())
		case n := <-nomChan:
			// we got a nomination from someone else, so reset our timer
			// in a way that doesn't race
			if !timer.Stop() {
				<-timer.C
			}
			timer.Reset(getCycleTime())
			fmt.Println("resetting timer because ", n)
		case <-signalChan:
			logger.Info("Received interrupt, stopping")
			return nil
		}
	}
}

func main() {
	var args struct {
		MinTime string `help:"minimum duration between postings [10m]"`
		MaxTime string `help:"maximum duration between postings [12m]"`
		ID      string `help:"nominator ID for logging"`
	}
	args.MaxTime = "12m"
	args.MinTime = "10m"
	arg.MustParse(&args)

	minTime, err := time.ParseDuration(args.MinTime)
	if err != nil {
		panic("Couldn't parse duration " + args.MinTime)
	}
	maxTime, err := time.ParseDuration(args.MaxTime)
	if err != nil {
		panic("Couldn't parse duration " + args.MaxTime)
	}
	if maxTime < minTime {
		panic("MaxTime must be more than MinTime")
	}

	baselogger := log.New()
	baselogger.Formatter = new(log.JSONFormatter)
	baselogger.Out = os.Stderr
	baselogger = honeycomb.Setup(baselogger)

	logger := baselogger.WithField("ID", args.ID)
	logger.WithField("mintime", args.MinTime).WithField("maxtime", args.MaxTime).Info("Starting nominator")
	txer := &dummyTxer{timing: time.Duration(15 * time.Second)}
	err = run(txer, minTime, maxTime, logger)
	if err != nil {
		logger.Fatalln(err)
	}
	os.Exit(0)
}
