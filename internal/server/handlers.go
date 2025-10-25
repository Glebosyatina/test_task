package server

import (
	"fmt"
	"log"
	"path"
	"strconv"
	"net/http"
	"encoding/json"
)

func (s *Server) Greet(w http.ResponseWriter, r *http.Request){
	fmt.Fprintf(w, "hello there\n")
}

//метод для получения информации обо всех подписках
func (s *Server) GetAllSubscriptions(w http.ResponseWriter,  r *http.Request){

	subs := make([]*Subscription, 0)
	
	rows, err := s.dbConn.Query("SELECT * FROM subscriptions")
	if err != nil{
		panic(err)
	}
	defer rows.Close()

	for rows.Next(){
		s := new(Subscription)
		
		err := rows.Scan(&s.Id, &s.NameService, &s.Price, &s.UserId, &s.StartDate, &s.EndDate)
		if err != nil{
			panic(err)
		}
		fmt.Println(*s)
		subs = append(subs, s)
	}
	
	
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

func (s *Server) RemoveSub(w http.ResponseWriter, r *http.Request){
	if r.Method != http.MethodDelete{
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}	
	
	id, _ := strconv.Atoi(path.Base(r.URL.Path))	

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




