package main

import (
	"fmt"
	"log"
	"os"
	"net/http"
	"net/url"
	"github.com/skip2/go-qrcode"
	"github.com/joho/godotenv"
)

func main() {
	http.HandleFunc("/buy", buyTicketHandler)
	http.HandleFunc("/qr",QRHandle)
	http.HandleFunc("/qrlove",QR1Handle)
	log.Println("Starting on port 8080   3...2...1...")
	log.Fatal(http.ListenAndServe(":8080", nil))

}
func getEnv(value *string, key string){
	err := godotenv.Load(".env")
	if err != nil{
		log.Fatal("не удалось загрузить .env", err)
		return
	}
	*value=os.Getenv(key)
}

func	QRHandle(w http.ResponseWriter, r *http.Request){
	var userURL string
	getEnv(&userURL, "USER_URL")
	name := r.URL.Query().Get("name")
	if name == ""{
	log.Println("300",r.URL)
	w.Write([]byte("введите имя"))
	return	
	}
	log.Println("200", r.URL)
	params := url.Values{}
	params.Set("name", name)
	urlString := fmt.Sprintf("%s?%s",userURL, params.Encode())
	png , err :=  qrcode.Encode(urlString, qrcode.Medium, 256)
	if err != nil{
	return 
	}
	w.Write(png)
}
func	QR1Handle(w http.ResponseWriter, r *http.Request){
	png , err :=  qrcode.Encode("Самый вредный котенок мой, моя любовь к тебе безгранична, поэтому я буду жертвовать для тебя все, что возможно", qrcode.Medium, 256)
	if err != nil{
	return 
	}
	w.Write(png)
}
func buyTicketHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "<h1>Покупка билета")	
	http.ServeFile(w,r,"public/buyTickets.html")
    // Здесь можно обработать создание билета и сохранить информацию в базе данных
}
