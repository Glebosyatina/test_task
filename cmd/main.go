package main

import(
	"log"

	"github.com/Glebosyatina/test_task/internal/server"
)

func main(){
	serv, err := server.NewServer()
	if err != nil{
		log.Fatal("Не удалось создать сервер: ", err)
	}
	err = serv.Run()
	if err != nil {
		log.Fatal("Ошибка при запуске сервиса: ", err)
	}
}
