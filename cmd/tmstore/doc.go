// Copyright 2017 Stratumn SAS. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

// The command tmstore starts an HTTP server with a tmstore.
//
// Usage:
//
//	$ tmstore -h
//	Usage of tmstore:
//	  -endpoint string
//     		endpoint used to communicate with Tendermint Core (default "tcp://127.0.0.1:46657")
//	  -tmWsRetryInterval duration
//		interval between tendermint websocket connection tries (default 5s)
//	  -http string
//	    	HTTP address (default ":5000")
//	  -maxmsgsize int
//	    	Maximum size of a received web socket message (default 32768)
//	  -tlscert string
//	    	TLS certificate file
//	  -tlskey string
//	    	TLS private key file
//	  -wspinginterval duration
//	    	Interval between web socket pings (default 54s)
//	  -wspongtimeout duration
//	    	Timeout for a web socket expected pong (default 1m0s)
//	  -wsreadbufsize int
//	    	Web socket read buffer size (default 1024)
//	  -wswritebufsize int
//	    	Web socket write buffer size (default 1024)
//	  -wswritechansize int
//	    	Size of a web socket connection write channel (default 256)
//	  -wswritetimeout duration
//	    	Timeout for a web socket write (default 10s)
//
// Docker
//
// A Docker image is available. To create a container, run:
//
//	$ docker run -p 5000:5000 stratumn/tmstore
package main