package main

import "fmt"

func main() {
	args := argparse()

	fmt.Println(args["path"])
	fmt.Println(args["exec"])
}
