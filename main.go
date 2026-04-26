package main

import "GoMusic/handler"

func main() {
	newServer().Spin()
}

func newServer() interface {
	Spin()
} {
	return handler.NewRouter()
}
