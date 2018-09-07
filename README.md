# nominator
This is a standalone web service that generates a random number on a regular basis and posts it to the ndau chain to
implement the node nomination transactions.

It is designed to be deployed with multiple instances, preferably spread around the world, so that in general
one of the instances will be successful at posting a transaction at the appropriate time.

Each instance is given a minimum and maximum duration between nominations. It sets a timer to wake up at a random
moment between those times.

It is also listenening to the ndau blockchain and whenever a node nomination transaction is posted, it resets the countdown to
a new random value.

If the countdown expires without a nomination transaction having been posted, it will generate and post one, then
reset its countdown again.

(Should two transactions arrive nearly simultaneously, the blockchain will reject the later one.)

This means that in general, each nominator will win the right to post a transaction 1/N tries, where N is the number of
nominators in existence.
