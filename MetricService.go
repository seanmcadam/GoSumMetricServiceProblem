
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

var Key = make(map[string]map[int64]int)

const Debugging bool = true
const TimeWindowSec int64 = 10
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
	return strings.Join( []string{ "{ ", valuestring, ": ", strconv.Itoa( returnval ), " }"}, "" )
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

	key_cleanup()
	return "{}"
} 


// -----------------------------------------------------------------------------
//
//
// -----------------------------------------------------------------------------
//
// It is better to run one at a time...
//
func auto_key_cleanup () {

	for true {
		time.Sleep( RunKeyCleanSeconds )
		if Debugging { fmt.Println( "auto_key_cleanup running" ) }
		key_cleanup()
	}

}

// -----------------------------------------------------------------------------
//
//
// -----------------------------------------------------------------------------
func key_cleanup () {

	TimeWindow := get_now_nsec() - TimeWindowNsec

	if Debugging { fmt.Println( "key_cleanup" ) }

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

	if Debugging {
		for k, tmap := range Key {
			fmt.Println( "PrintKey ", k );
			for t, v := range tmap {
			fmt.Println( "\t", t, ": ", v );
			}
		}
	}
}
