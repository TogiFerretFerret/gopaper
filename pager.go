package main
import (
   "log"
   "net/http"
   "fmt"
)

func main() {
	_, err := http.Get("http://localhost:6969/page")
	if err != nil {
   		log.Fatalln(err)
	}
	fmt.Println("Paged!")
}
