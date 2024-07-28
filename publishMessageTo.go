package main

import (
	//"log"
)

/*
publishMessageTo {{{
*/
func publishMessageTo(args []string, cc confClass) error {
	if err := publishMessage(args, cc); err != nil {
		return(err)
	}
	return nil
}

// }}}

