// Package cli provides Command-line interface functions for skancoin
package cli

import (
	"flag"
	"fmt"
	"os"
	"runtime"

	"github.com/skan0779/skan_coin_2022/explorer"
	"github.com/skan0779/skan_coin_2022/rest"
)

func Start() {
	if len(os.Args) == 1 {
		fmt.Printf("Welcome to Skan Coin \n\n")
		fmt.Printf("Please use the following commands \n\n")
		fmt.Printf("-mode:	set the server mode between 'rest' and 'html' \n")
		fmt.Printf("-port:	set the server port number  \n")
		os.Exit(0)
	}

	mode := flag.String("mode", "rest", "Set the mode of server | default: rest api")
	port := flag.Int("port", 4000, "Set the port number of server | default: 4000")
	flag.Parse()
	switch *mode {
	case "html":
		explorer.Start(*port)
	case "rest":
		rest.Start(*port)
	default:
		fmt.Printf("Welcome to Skan Coin \n\n")
		fmt.Printf("Please use the following commands \n\n")
		fmt.Printf("-mode:	set the server mode between 'rest' and 'html' \n")
		fmt.Printf("-port:	set the server port number  \n")
		// run the main()'s defer before the exit
		runtime.Goexit()
	}
}
