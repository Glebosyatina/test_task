package server

import (
	"net/http"
	"errors"
	"database/sql"
	_ "github.com/lib/pq"
	"fmt"

	"github.com/Glebosyatina/test_task/config"
)

type Server struct{
	Addr string
	dbConn *sql.DB
}

func NewServer() (*Server, error){

	//парсим конфиг
	conf := config.ReadConfig()

	//заполняем структуру сервер
	connInfo := fmt.Sprintf("user=%s dbname=%s sslmode=disable password=%s", conf.DbConf.DbUser, conf.DbConf.DbName, conf.DbConf.DbPassword)
	conn, err := sql.Open("postgres", connInfo)
	if err != nil{
		return nil, errors.New("Не удалось открыть соединение с бд")
	}

	server := &Server{
		Addr: conf.Server.Addr,
		dbConn: conn,
	}

	return server, nil
}

func (s *Server) Run() error {

	http.HandleFunc("/", s.Greet)
	http.HandleFunc("/subs", s.GetAllSubscriptions)
	http.HandleFunc("/subs/add", s.CreateSub)
	http.HandleFunc("/subs/rm/", s.RemoveSub)



	err := http.ListenAndServe(s.Addr, nil)	
	if err == nil{
		return nil
	} else {
		return errors.New("Не удалось запустить сервер")
	}
}


