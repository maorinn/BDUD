package main

import (
	"os"
)

func main() {
	err:=os.Rename("./download/台灣SWAG.mp4.mp4","/root/googleDrive/台灣SWAG.mp4.mp4")
	if err != nil {
		panic(err)
	}
}
