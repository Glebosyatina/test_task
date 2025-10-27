package server

import (
	"net/http"
	"errors"
	"log"
	"os"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"fmt"

	"github.com/Glebosyatina/test_task/config"
)

type Server struct{
	Addr string
	dbConn *sqlx.DB
	Logs Loggers
}
type Loggers struct{
	InfoLog *log.Logger
	ErrorLog *log.Logger
}

func NewServer() (*Server, error){

	//парсим конфиг
	conf := config.ReadConfig()
	//заполняем структуру сервер
	connInfo := fmt.Sprintf("host=%s user=%s dbname=%s sslmode=disable password=%s", conf.DbConf.Host, conf.DbConf.DbUser, conf.DbConf.DbName, conf.DbConf.DbPassword)
	conn, err := sqlx.Open("postgres", connInfo)
	if err != nil{
		return nil, errors.New("Не удалось открыть соединение с бд")
	}
		

	server := &Server{
		Addr: conf.Server.Addr,
		dbConn: conn,
	}
	
	//миграции
	server.dbConn.MustExec("DROP TABLE IF EXISTS subscriptions")
	server.dbConn.MustExec(`CREATE TABLE subscriptions (
		id SERIAL    PRIMARY KEY,
		service_name VARCHAR,
		price        INTEGER,
		user_id		 UUID,
		start_date	 VARCHAR,
		end_date	 VARCHAR
	)`)

	//логирование
	file, err := os.OpenFile("app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil{
		log.Fatal("Не удалось открыть файл: ", err)	
	}

	server.Logs.InfoLog = log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	server.Logs.ErrorLog = log.New(os.Stdout, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)
	server.Logs.InfoLog.SetOutput(file)
	server.Logs.ErrorLog.SetOutput(file)

	return server, nil
}

func (s *Server) Run() error {

	http.HandleFunc("/", s.Greet)
	http.HandleFunc("/subs", s.GetAllSubscriptions)
	http.HandleFunc("/subs/add", s.CreateSub)
	http.HandleFunc("/subs/rm/", s.RemoveSub)
	http.HandleFunc("/subs/up/", s.UpdateSub)
	http.HandleFunc("/subs/sum", s.GetSumSubs)
	

	s.Logs.InfoLog.Println("Сервер запущен на", s.Addr, "порту")
	err := http.ListenAndServe(s.Addr, nil)	
	if err != nil{
		return errors.New("Не удалось запустить сервер")
	}

	return nil
}


