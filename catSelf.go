package main

/*
catSelf
*/
func catSelf(args []string, cc confClass) error {
	if err := getNote(args, cc); err != nil {
		return err
	}
	return nil
}

//
