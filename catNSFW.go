package main

/*
catNSFW {{{
*/
func catNSFW(args []string, cc confClass) error {
	if err := getNote(args, cc); err != nil {
		return err
	}
	return nil
}

// }}}
