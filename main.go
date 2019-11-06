package main

import (
	"fmt"

	server2 "github.com/abdiUNO/featherr/server"
)

func main() {

	s, err := server2.NewServer()

	if err != nil {
		fmt.Print(err)
	}

	err = s.ListenAndServe()

	fmt.Println("Listing on localhost")

	if err != nil {
		fmt.Println(err)
	}
}
