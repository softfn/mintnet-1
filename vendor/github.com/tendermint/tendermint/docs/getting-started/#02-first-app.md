# First Tendermint App

As a general purpose blockchain engine, Tendermint is agnostic to the application you want to run.
So, to run a complete blockchain that does something useful, you must start two programs:
one is Tendermint Core, the other is your application, which can be written in any programming language.
Recall from [the intro to ABCI](/intro/abci-overview) that Tendermint Core handles all the p2p and consensus stuff,
and just forwards transactions to the application when they need to be validated, or when they're ready to be committed to a block.

In this guide, we show you some examples of how to run an application using Tendermint.

**Note:** It is highly recommended to read the [Using Tendermint Guide](/docs/guides/using-tendermint) prior to working through this tutorial.

## Install

First, make sure you have [installed Tendermint](/download).
The first apps we will work with are written in Go. 
To install them, you need to [install Go](https://golang.org/doc/install) and 
[put `$GOPATH/bin` in your `$PATH`](https://github.com/tendermint/tendermint/wiki/Setting-GOPATH). 

Then run

```
go get -u github.com/tendermint/abci/cmd/...
```

If there is an error, install and run the `glide` tool to pin the dependencies:

```
go get github.com/Masterminds/glide
cd $GOPATH/src/github.com/tendermint/abci
glide install
go install ./cmd/...
```

Now you should have the `abci-cli` plus two apps installed: 

```
dummy --help
counter --help
```

These binaries are installed on `$GOPATH/bin` and all come from within the `./cmd/...` directory of the abci repository.

Both of these example applications are in Go. See below for an application written in Javascript.

Now, let's run some apps!

## A First Example - Dummy

The dummy app is a [Merkle tree](https://en.wikipedia.org/wiki/Merkle_tree) that just stores all transactions.
If the transaction contains an `=`, eg. `key=value`, 
then the `value` is stored under the `key` in the Merkle tree.
Otherwise, the full transaction bytes are stored as the key and the value.

Let's start a dummy application.

```
dummy
```

In another terminal, we can start Tendermint.
If you have never run Tendermint before, use:

```
tendermint init 
tendermint node
```

If you have used Tendermint, you may want to reset the data for a new blockchain by running `tendermint unsafe_reset_all`.
Then you can run `tendermint node` to start Tendermint, and connect to the app.
For more details, see [the guide on using Tendermint](/docs/guides/using-tendermint).

You should see Tendermint making blocks! 
We can get the status of our Tendermint node as follows:

```
curl -s localhost:46657/status
```

The `-s` just silences `curl`. For nicer output, pipe the result into a tool like [jq](https://stedolan.github.io/jq/) 
or [jsonpp](https://github.com/jmhodges/jsonpp).

Now let's send some transactions to the dummy.

```
curl -s 'localhost:46657/broadcast_tx_commit?tx="abcd"'
```

Note the single quote (`'`) around the url, which ensures that the double quotes (`"`) are not escaped by bash.
This command sent a transaction with bytes `abcd`, so `abcd` will be stored as both the key and the value in the Merkle tree.
The response should look something like:

```
{"jsonrpc":"2.0","id":"","result":[98,{"check_tx":{},"deliver_tx":{}}],"error":""}
```

The `98` is a type-byte, and can be ignored (it's useful for serializing and deserializing arbitrary json).
Otherwise, this result is empty - there's nothing to report on and everything is OK.

We can confirm that our transaction worked and the value got stored by querying the app:

```
curl -s 'localhost:46657/abci_query?data="abcd"&path=""&prove=false'
```

The `path` and `prove` arguments can be ignored for now, and in a future release can be left out.
The result should look like:


```
{"jsonrpc":"2.0","id":"","result":[112,{"response":{"value":"61626364","log":"exists"}}],"error":""}
```

Again, the `112` is the type-byte. Note the `value` in the result (`61626364`); this is the hex-encoding of the ASCII of `abcd`.
You can verify this in a python shell by running `"61626364".decode('hex')`.
Stay tuned for a future release that makes this output more human-readable ;). 

Now let's try setting a different key and value:

```
curl -s 'localhost:46657/broadcast_tx_commit?tx="name=satoshi"'
```

Now if we query for `name`, we should get `satoshi`, or `7361746F736869` in hex:

```
curl -s 'localhost:46657/abci_query?data="name"&path=""&prove=false'
```

Try some other transactions and queries to make sure everything is working!

## Another Example - Counter

Now that we've got the hang of it, let's try another application, the "counter" app.

The counter app doesn't use a Merkle tree, it just counts how many times we've sent a transaction,
or committed the state. 

This application has two modes: `serial=off` and `serial=on`.

When `serial=on`, transactions must be a big-endian encoded incrementing integer, starting at 0.

If `serial=off`, there are no restrictions on transactions.

In a live blockchain, transactions collect in memory before they are committed into blocks.
To avoid wasting resources on invalid transactions,
ABCI provides the `CheckTx` message,
which application developers can use to accept or reject transactions,
before they are stored in memory or gossipped to other peers.

In this instance of the counter app, with `serial=on`, `CheckTx` only allows transactions whose integer is greater than the last committed one.

Let's kill the previous instance of `tendermint` and the `dummy` application, and start the counter app.
We can enable `serial=on` with a flag:

```
counter --serial
```

In another window, reset then start Tendermint:

```
tendermint unsafe_reset_all
tendermint node
```

Once again, you can see the blocks streaming by. Let's send some transactions.
Since we have set `serial=on`, the first transaction must be the number `0`:

```
curl localhost:46657/broadcast_tx_commit?tx=0x00
```

Note the empty (hence successful) response.
The next transaction must be the number `1`. If instead, we try to send a `5`, we get an error:

```
> curl localhost:46657/broadcast_tx_commit?tx=0x05
{"jsonrpc":"2.0","id":"","result":[98,{"check_tx":{},"deliver_tx":{"code":3,"log":"Invalid nonce. Expected 1, got 5"}}],"error":""}
```

But if we send a `1`, it works again:

```
> curl localhost:46657/broadcast_tx_commit?tx=0x01
{"jsonrpc":"2.0","id":"","result":[98,{"check_tx":{},"deliver_tx":{}}],"error":""}
```

For more details on the `broadcast_tx` API, 
see [the guide on using Tendermint](/docs/guides/using-tendermint).

## Example in Another Language - CounterJS

We also want to run applications in another language - in this case, we'll run a Javascript version of the `counter`.
To run it, you'll need to [install node](https://nodejs.org/en/download/).

You'll also need to fetch the relevant repository, from https://github.com/tendermint/js-abci then install it.
As go devs, we keep all our code under the `$GOPATH`, so run:

```
go get github.com/tendermint/js-abci &> /dev/null
cd $GOPATH/src/github.com/tendermint/js-abci/example
npm install
```

Kill the previous `counter` and `tendermint` processes. Now run the app:

```
node example/app.js
```

In another window, reset and start `tendermint`:

```
tendermint unsafe_reset_all
tendermint node
```

Once again, you should see blocks streaming by - but now, our application is written in javascript!
Try sending some transactions, and like before - the results should be the same:

```
curl localhost:46657/broadcast_tx_commit?tx=0x00 # ok
curl localhost:46657/broadcast_tx_commit?tx=0x05 # invalid nonce
curl localhost:46657/broadcast_tx_commit?tx=0x01 # ok
```

Neat, eh?

## A More Interesting Example - Basecoin

Before concluding, we'd like to introduce you to our star application, [Basecoin](https://github.com/tendermint/basecoin).
Unlike the `dummy` and `counter`, which are strictly for example purposes, 
`basecoin` is designed to be actually useful - it's a general purpose framework for building cryptocurrencies.

The default `basecoin` application is a multi-asset cryptocurrency that supports inter-blockchain communication.
For more details on how basecoin works and how to use it, see our [basecoin guide](https://github.com/tendermint/basecoin/blob/develop/docs/guide/basecoin-basics.md)

## Next Step

In this tutorial you learned how to run applications using Tendermint on a single node.
You saw how applications could be written in different languages, 
and how to send transactions and query for the latest state.
But the true power of Tendermint comes from its ability to securely and efficiently run an application 
across a distributed network of nodes, while keeping them all in sync using its state-of-the-art consensus protocol.
This is the subject of the next tutorial, where we show you [how to deploy Tendermint networks](/docs/getting-started/deploy-testnet).
