package main

import "flag"

func argparse() map[string]string {
	args := make(map[string]string)

	// parse the args
	pathFlag := flag.String("p", "", "path to watch")
	execFlag := flag.String("c", "", "command to execute")

	// parse the flags
	flag.Parse()

	// add to map
	args["path"] = *pathFlag
	args["exec"] = *execFlag

	// return map
	return args
}
