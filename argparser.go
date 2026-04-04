package main

import "flag"

func argparse() watchMeta {
	// parse the args
	pathFlag := flag.String("p", "", "path to watch")
	execFlag := flag.String("c", "", "command to execute")

	// parse the flags
	flag.Parse()

	return watchMeta{
		path: *pathFlag,
		cmd: *execFlag,
	}
}
