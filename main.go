package main

import "hotkeysdouban/douban"

func main() {
	loop := douban.NewMainLoop()
	loop.Start()
}
