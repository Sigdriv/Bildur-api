package main

import "github.com/Sigdriv/Bildur-api/handler"

func main() {
	srv := handler.CreateHandler()

	srv.CreateGinGroup()
}
