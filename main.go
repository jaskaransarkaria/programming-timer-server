package main

import (
	"github.com/jaskaransarkaria/programming-timer-server/http-routes"
	"fmt"
	"net/http"
	"log"
	"flag"
)
// flag allows you to create cli flags and assign a default
var addr = flag.String("addr", "localhost:8080", "http service address")

func main() {
	fmt.Println("Golang WebSockets running...")
	httproutes.SetupRoutes()
	flag.Parse()
	fmt.Println("Listening on:", *addr)
	log.Fatal(http.ListenAndServe(*addr, nil))
}
