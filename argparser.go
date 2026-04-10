package main

import "flag"

// argparse parses the command-line flags and returns a watchMeta populated
// with the values provided by the user. Unset flags are left as empty strings;
// watcher applies defaults (e.g. CWD for -p) when it processes the result.
func argparse() watchMeta {
	// parse the args
	pathFlag := flag.String("p", "", "path to watch")
	execFlag := flag.String("c", "", "command to execute")

	// parse the flags
	flag.Parse()

	return watchMeta{
		path: *pathFlag,
		cmd:  *execFlag,
	}
}
