package main

import "GoMusic/handler"

func main() {
	r := handler.NewRouter()
	r.Spin()
}
