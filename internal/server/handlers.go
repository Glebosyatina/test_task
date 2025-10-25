package server

import (
	"fmt"
	"log"
	"path"
	"strconv"
	"net/http"
	"encoding/json"

	"github.com/google/uuid"
)

func (s *Server) Greet(w http.ResponseWriter, r *http.Request){
	fmt.Fprintf(w, "hello there\n")
}

//метод для получения информации обо всех подписках
func (s *Server) GetAllSubscriptions(w http.ResponseWriter,  r *http.Request){
	if r.Method != http.MethodGet{
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}	

	//слайс подписок в который будем читать из базы
	subs := make([]*Subscription, 0)
	
	rows, err := s.dbConn.Query("SELECT * FROM subscriptions")
	if err != nil{
		panic(err)
	}
	defer rows.Close()
	
	//проходим по каждой записи и добавляем в слайс
	for rows.Next(){
		s := new(Subscription)
		
		err := rows.Scan(&s.Id, &s.NameService, &s.Price, &s.UserId, &s.StartDate, &s.EndDate)
		if err != nil{
			panic(err)
		}
		fmt.Println(*s)
		subs = append(subs, s)
	}
	
	//пишем ответ в json формате
	json.NewEncoder(w).Encode(subs)
}


//метод создания подписки
func (s *Server) CreateSub(w http.ResponseWriter, r *http.Request){
	if r.Method != http.MethodPost{
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		log.Fatal("Подписка не создана, недопустимый метод")
	}

	var sub Subscription
	//парсим тело запроса в структуру	
	err := json.NewDecoder(r.Body).Decode(&sub)
	if err != nil{
		http.Error(w, "Подписка не создалась", http.StatusInternalServerError)
		log.Fatal("Не удалось создать подписку: ", err.Error())
	}
	fmt.Println(sub)
	
	//вставляем запись в бд
	result, err := s.dbConn.Exec("INSERT INTO subscriptions (service_name, price, user_id, start_date, end_date) VALUES ($1, $2, $3, $4, $5)", sub.NameService, sub.Price, sub.UserId, sub.StartDate, sub.EndDate)
	if err != nil{
		http.Error(w, "Ошибка выполнения запроса к бд", http.StatusInternalServerError)
		log.Fatal("Не удалось выполнить запрос по вставке записи в бд")
	}

	subsAdded, err := result.RowsAffected()
	if err != nil{
		http.Error(w, "Ошибка получения количества вставленных записей в бд", http.StatusInternalServerError)
		log.Fatal("Не удалось выполнить запрос к бд")
	}
	fmt.Fprintf(w, "В базу добавлена %d подписка", subsAdded)
}

//метод удаления подписки
func (s *Server) RemoveSub(w http.ResponseWriter, r *http.Request){
	//проверка на метод запроса
	if r.Method != http.MethodDelete{
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}	
	//находим id в url	
	id, _ := strconv.Atoi(path.Base(r.URL.Path))	
	//удаляем запись из базы
	result, err := s.dbConn.Exec("DELETE FROM subscriptions WHERE id=$1", id)
	if err != nil{
		http.Error(w, "Ошибка выполнения запроса к бд", http.StatusInternalServerError)
		log.Fatal("Не удалось выполнить запрос к бд")
	}
	subsRemoved, err := result.RowsAffected()
	if err != nil {
		http.Error(w, "Ошибка получения количества удаленных записей", http.StatusInternalServerError)
		log.Fatal("Не удалось выполнить запрос к бд")
	}
	
	fmt.Fprintf(w, "Из базы удалена %d подписка", subsRemoved)
}


//метод обновления подписки
func (s *Server) UpdateSub(w http.ResponseWriter, r *http.Request){
	if r.Method != http.MethodPut{
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	id, _ := strconv.Atoi(path.Base(r.URL.Path))
	
	//парсим тело запроса в структуру
	var sub Subscription
	err := json.NewDecoder(r.Body).Decode(&sub)
	if err != nil{
		http.Error(w, "не удалось распарсить тело запроса", http.StatusBadRequest)
		return
	}
	
	//обновляем запись в бд
	res, err := s.dbConn.Exec("UPDATE subscriptions SET service_name=$1, price=$2, user_id=$3, start_date=$4, end_date=$5 WHERE id=$6", sub.NameService, sub.Price, sub.UserId, sub.StartDate, sub.EndDate, id)
	if err != nil{
		http.Error(w, "ошибка выполнения запроса к бд", http.StatusInternalServerError)
		return
	}

	subsUpdated, err := res.RowsAffected()
	if err != nil{
		http.Error(w, "Ошибка обновления записи в бд", http.StatusInternalServerError)
		log.Fatal("Не удалось выполнить запрос к бд")
	}

	fmt.Fprintf(w, "В базе обновлена %d подписка", subsUpdated)
	
}

//метод получения суммарной стоимости подписок за перид с группировкой по user_id и service_name
func (s *Server) GetSumSubs(w http.ResponseWriter, r *http.Request){
	// localhost:port/subs/sum?start&end
	//вытаскиваем из url параметры start, end для выбранного периода вида "2025-10-24"	
	start := r.URL.Query().Get("start")
	end := r.URL.Query().Get("end")

	
	//делаем выборку по пользователю и названию подписки, с подсчетом общей стоимости подписок
	rows, err := s.dbConn.Query("SELECT user_id, service_name, sum(price) FROM subscriptions WHERE start_date >= $1 AND end_date <= $2 GROUP BY user_id, service_name", start, end)
	if err != nil{
		http.Error(w, "ошибка выполнения запроса к бд", http.StatusInternalServerError)
	}		
	defer rows.Close()	

	//структура результирующих сумм по сгруппированным и отфильтрованным подпискам
	type RecordSum struct{
		ServiceName string		`json:"service_name"`
		UserId		uuid.UUID	`json:"user_id"`
		Sum			uint		`json:"sum"`
	}

	sums := make([]*RecordSum, 0)	
		
	for rows.Next(){
		s := new(RecordSum)

		err := rows.Scan(&s.UserId, &s.ServiceName, &s.Sum)
		if err != nil{
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}
		fmt.Println(*s)
		sums = append(sums, s)	
	}

	//пишем слайс в тело ответа 
	json.NewEncoder(w).Encode(sums)
}




