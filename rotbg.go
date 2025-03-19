package main
import (
   "log"
   "net/http"
   "fmt"
)

func main() {
	_, err := http.Get("http://localhost:6969/change_wallpaper")
	if err != nil {
   		log.Fatalln(err)
	}
	fmt.Println("OK")
}
