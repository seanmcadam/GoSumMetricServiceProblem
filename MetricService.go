
package main
    
import (
	"fmt"
	"strconv"
	"strings"
	"time"
	"github.com/hoisie/web"
)
    
//
// Global storage variable [string][time][value]
// Key-string, nsec-int64, value
//
// This assumes that two calls will not be made on the same nanosecond, good enough for govt work...
//
// Keys are set to live for 3600 seconds (1 Hour)
// RunKeyCleanupSeconds runs 5 seconds after the keys should expire, always garunteeing that a key does not happen the be deleted early.
//
//
// Global Data Storage for the key and time stamps
//
var Key = make(map[string]map[int64]int)

const Debugging bool = false
const TimeWindowSec int64 = 3600
const TimeWindowNsec int64 = TimeWindowSec * 1000000000
const RunKeyCleanSeconds time.Duration = ( time.Duration(TimeWindowSec) * time.Second ) + 5;

func main() {

	// Start the auto Key cleanup function
	go auto_key_cleanup()

	// Required for the problem
	web.Post("/metric/(.*)/", metric_post_handle)
	web.Post("/metric/(.*)", metric_post_handle)
	web.Get("/metric/(.*)/sum", metric_get_handle)

	//
	// Extra handlers to neaten things up and help with debugging
	//
	web.Post("/print", key_print_handle)
	web.Get("/print", key_print_handle)
	web.Post("/cleanup", key_cleanup_handle)
	web.Get("/cleanup", key_cleanup_handle)
	web.Post("(.*)", bad_uri_handle)
	web.Get("(.*)", bad_uri_handle)
	web.Get("/(.*)", bad_uri_handle)

	//
	// Run until a signal is recieved
	//
	fmt.Println("Starting Service")
	web.Run("0.0.0.0:9999")
}


// -----------------------------------------------------------------------------
//
//
// -----------------------------------------------------------------------------
func metric_get_handle( ctx *web.Context, key string) string { 
	returnval := 0;

	nsec := get_now_nsec()

	TimeWindow := nsec - TimeWindowNsec

	if Debugging {
		fmt.Println("Key:", key)
		fmt.Println("      NSEC:", nsec)
		fmt.Println("TimeWindow:", TimeWindow)
	}

	//
	// Check to see if the key exists
	// then read each of the time entries
	//
	if timeMap, ok := Key[key]; ok  {

		for k, v := range timeMap {
			if( k > TimeWindow ) {
				returnval += v
				if Debugging { fmt.Println("adding value:", v) }
			}	
		}
	} else {
		if Debugging { fmt.Println("Emptykey: ", key) }
	}

	if Debugging { print_keys() }

	valuestring := strconv.Quote("value")
	return strings.Join( []string{ "{\n  ", valuestring, ": ", strconv.Itoa( returnval ), "\n}"}, "" )
} 


// -----------------------------------------------------------------------------
//
// POST of a key with a value.
//
// -----------------------------------------------------------------------------
func metric_post_handle( ctx *web.Context, key string) string { 
	var value = 0;

	nsec := get_now_nsec()

	// fmt.Println("NSEC:", nsec)

	//
	// get the value in the POST
	// ctx.Params is a map[string]string
	//
	if valueMap, ok := ctx.Params["value"]; ok  {

		// fmt.Println("valueMap:", valueMap)
		if val, err := strconv.Atoi( valueMap ); err == nil {
			value = val
		} else {
			// POST value non-numeric or error { "value": ??? } found, consider this an error and return a 404
			e := "POST value non-numeric or error"
			ctx.NotFound( e )
			return e
		}

	} else {
		e := "POST with out a { value: nnn }"
		ctx.NotFound( e );
		return e
	}


	// Check to see if the key exists
	if timeMap, ok := Key[key]; ok  {
		timeMap[nsec] = value;
	} else {
		var m = make(map[int64]int)
		m[nsec] = value
		Key[key] = m
	}

	if Debugging { print_keys() }

	// Just some fun stuff...
	ctx.SetHeader("X-Powered-By", "superduperwebmanagementool", true)
	ctx.SetHeader("Connection", "close", true)

	return "{}"
} 



// -----------------------------------------------------------------------------
//
//
// -----------------------------------------------------------------------------
func bad_uri_handle( ctx *web.Context, val string) string { 

	ctx.NotFound( "" )
	return strings.Join( []string{ "BAD URI - 404\n", "\t", val, "\n" }, "" )
} 

// -----------------------------------------------------------------------------
//
//
// -----------------------------------------------------------------------------
func key_cleanup_handle( ctx *web.Context) string { 

	go key_cleanup()
	return "{}"
} 

// -----------------------------------------------------------------------------
//
//
// -----------------------------------------------------------------------------
func key_print_handle( ctx *web.Context) string { 

	go print_keys()
	return "{}"
} 


// -----------------------------------------------------------------------------
//
//
// -----------------------------------------------------------------------------
func auto_key_cleanup () {

	for true {
		time.Sleep( RunKeyCleanSeconds )
		if Debugging { fmt.Println( "auto_key_cleanup running" ) }
		key_cleanup()
	}

}

// -----------------------------------------------------------------------------
//
// Future option would be to add in a counter of entries for each key and
// remove keys with 0 times stored in them. Not a huge efficiency thing unless
// there are going to be thousands upon thousands of keys accumulating...
// Would depend on the use case...
//
// -----------------------------------------------------------------------------
func key_cleanup () {

	//
	// Calculate the oldest time that should be kept
	//
	TimeWindow := get_now_nsec() - TimeWindowNsec

	if Debugging { fmt.Println( "key_cleanup" ) }

	// 
 	// itterate over each key, then itterate over each time and delete times that have expired
	//
	for k, tmap := range Key {
		for t, _ := range tmap {
			if( t < TimeWindow ) {
				delete(tmap, t)
				if Debugging { fmt.Println( "key_cleanup key:", k, " delete: ", t ) }
			}
		}
	}
}


// -----------------------------------------------------------------------------
//
//
// -----------------------------------------------------------------------------
func get_now_nsec() int64 {

	now := time.Now()
	nsec := now.UnixNano()
	return nsec
}


//
// Debugging functions
//
// -----------------------------------------------------------------------------
//
//
// -----------------------------------------------------------------------------
func print_keys () {

	for k, tmap := range Key {
		fmt.Println( "PrintKey ", k );
		for t, v := range tmap {
		fmt.Println( "\t", t, ": ", v );
		}
	}
}
