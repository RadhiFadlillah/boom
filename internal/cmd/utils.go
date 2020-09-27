package cmd

import (
	"os"
)

func panicError(err error, prefixes ...string) {
	if err != nil {
		for _, prefix := range prefixes {
			cError.Print(prefix + " ")
		}

		cError.Println(err)
		os.Exit(1)
	}
}
