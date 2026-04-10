package main

import "flag"

// argparse parses the command-line flags and returns a watchMeta populated
// with the values provided by the user. Unset flags are left as empty strings;
// watcher applies defaults (e.g. CWD for -p) when it processes the result.
//
// Recognised flags:
//
//	-p  root path to watch (default: current working directory)
//	-x  shell command to re-run on changes
//	-c  path to a YAML config file (alternative to -p / -x)
func argparse() watchMeta {
	// parse the args
	pathFlag := flag.String("p", "", "path to watch")
	execFlag := flag.String("x", "", "command to execute")
	configFlag := flag.String("c", "", "path to config file")

	// parse the flags
	flag.Parse()

	return watchMeta{
		path:   *pathFlag,
		cmd:    *execFlag,
		config: *configFlag,
	}
}
