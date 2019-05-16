package p3

import "net/http"

type Route struct {
	Name        string
	Method      string
	Pattern     string
	HandlerFunc http.HandlerFunc
}

type Routes []Route

var routes = Routes{
	Route{
		"Show",
		"GET",
		"/show",
		Show,
	},

	Route{
		"Upload",
		"POST",
		"/upload",
		Upload,
	},
	Route{
		"UploadBlock",
		"GET",
		"/block/{height}/{hash}",
		UploadBlock,
	},
	Route{
		"HeartBeatReceive",
		"POST",
		"/heartbeat/receive",
		HeartBeatReceive,
	},
	Route{
		"Start",
		"GET",
		"/start",
		Start,
	},
	// Add a route: Name is "Canonical", Method is GET, Pattern is "/canonical", HandlerFunc is "Canonical".
	Route{
		"Canonical",
		"GET",
		"/canonical",
		Canonical,
	},
	Route{
		"Transfer",
		"POST",
		"/transfer",
		Transfer,
	},
	Route{
		"TransactionReceive",
		"POST",
		"/transaction/receive",
		TransactionReceive,
	},
	Route{
		"MyBalance",
		"GET",
		"/mybalance",
		MyBalance,
	},
	Route{
		"MyTXs",
		"GET",
		"/mytxs",
		MyTXs,
	},
	Route{
		"AllTXs",
		"GET",
		"/txs",
		AllTXs,
	},
}