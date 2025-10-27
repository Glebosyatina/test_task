package server

import (
	"fmt"
	"path"
	"strconv"
	"net/http"
	"encoding/json"

	"github.com/google/uuid"
)

//метод для получения информации обо всех подписках
func (s *Server) GetAllSubscriptions(w http.ResponseWriter,  r *http.Request){
	if r.Method != http.MethodGet{
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		s.Logs.InfoLog.Println("Неверный метод для получения информации о подписках")
		return
	}	

	//слайс подписок в который будем читать из базы
	subs := make([]*Subscription, 0)
	
	rows, err := s.dbConn.Query("SELECT * FROM subscriptions")
	if err != nil{
		http.Error(w, "Не удалось получить информацию о подписках", http.StatusInternalServerError)
		s.Logs.ErrorLog.Println("Не удалось выполнить запрос к бд")
		return
	}
	defer rows.Close()
	
	//проходим по каждой записи и добавляем в слайс
	for rows.Next(){
		sub := new(Subscription)
		
		err := rows.Scan(&sub.Id, &sub.NameService, &sub.Price, &sub.UserId, &sub.StartDate, &sub.EndDate)
		if err != nil{
			s.Logs.ErrorLog.Println("Не удалось распарсить запись из бд")
		}
		subs = append(subs, sub)
	}
	
	//пишем ответ в json формате
	json.NewEncoder(w).Encode(subs)
}


//метод создания подписки
func (s *Server) CreateSub(w http.ResponseWriter, r *http.Request){
	if r.Method != http.MethodPost{
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		s.Logs.ErrorLog.Println("Подписка не создана, недопустимый метод")
		return
	}

	var sub Subscription
	//парсим тело запроса в структуру	
	err := json.NewDecoder(r.Body).Decode(&sub)
	if err != nil{
		http.Error(w, "Не удалось создать подписку", http.StatusInternalServerError)
		s.Logs.ErrorLog.Println("Не удалось распарсить тело запроса в подписку: ", err.Error())
		return
	}
	
	//вставляем запись в бд
	result, err := s.dbConn.Exec("INSERT INTO subscriptions (service_name, price, user_id, start_date, end_date) VALUES ($1, $2, $3, $4, $5)", sub.NameService, sub.Price, sub.UserId, sub.StartDate, sub.EndDate)
	if err != nil{
		http.Error(w, "Не удалось создать подписку", http.StatusInternalServerError)
		s.Logs.ErrorLog.Println("Не удалось выполнить запрос по вставке записи в бд")
	}

	subsAdded, err := result.RowsAffected()
	if err != nil{
		http.Error(w, "Не удалось создать подписку", http.StatusInternalServerError)
		s.Logs.ErrorLog.Println("Не удалось выполнить запрос к бд")
		return
	}
	fmt.Fprintf(w, "В базу добавлена %d подписка", subsAdded)
}

//метод получения информации об одной подписке по id
func (s *Server) GetSub(w http.ResponseWriter, r *http.Request){
	//проверка на метод запроса
	if r.Method != http.MethodGet{
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		s.Logs.InfoLog.Println("Недопустимый метод запроса")
		return
	}	
	//находим id в url	
	id, _ := strconv.Atoi(path.Base(r.URL.Path))	
	sub := new(Subscription)
	//запрос к бд, парсим в структуру 
	row := s.dbConn.QueryRow("SELECT * FROM subscriptions WHERE id=$1", id)
	err := row.Scan(&sub.Id, &sub.NameService, &sub.Price, &sub.UserId, &sub.StartDate, &sub.EndDate)

	if err != nil{
		http.Error(w, "Не удалось получить информацию о подписке", http.StatusInternalServerError)
		s.Logs.ErrorLog.Println("Не удалось выполнить запрос к бд")
		return
	}

	json.NewEncoder(w).Encode(sub)	
}


//метод удаления подписки
func (s *Server) RemoveSub(w http.ResponseWriter, r *http.Request){
	//проверка на метод запроса
	if r.Method != http.MethodDelete{
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		s.Logs.InfoLog.Println("Недопустимый метод запроса")
		return
	}	
	//находим id в url	
	id, _ := strconv.Atoi(path.Base(r.URL.Path))	
	//удаляем запись из базы
	result, err := s.dbConn.Exec("DELETE FROM subscriptions WHERE id=$1", id)
	if err != nil{
		http.Error(w, "Не удалось удалить подписку", http.StatusInternalServerError)
		s.Logs.ErrorLog.Println("Не удалось выполнить запрос к бд")
		return
	}
	subsRemoved, err := result.RowsAffected()
	if err != nil {
		http.Error(w, "Не удалось удалить подписку", http.StatusInternalServerError)
		s.Logs.ErrorLog.Println("Не удалось выполнить запрос к бд")
		return
	}
	
	fmt.Fprintf(w, "Из базы удалена %d подписка\n", subsRemoved)
}


//метод обновления подписки
func (s *Server) UpdateSub(w http.ResponseWriter, r *http.Request){
	if r.Method != http.MethodPut{
		http.Error(w, "Недопустимый метод запроса", http.StatusMethodNotAllowed)
		return
	}

	id, _ := strconv.Atoi(path.Base(r.URL.Path))
	
	//парсим тело запроса в структуру
	var sub Subscription
	err := json.NewDecoder(r.Body).Decode(&sub)
	if err != nil{
		http.Error(w, "не удалось распарсить тело запроса", http.StatusBadRequest)
		s.Logs.ErrorLog.Println("Не удалось паспарсить тело запроса")
		return
	}
	
	//обновляем запись в бд
	res, err := s.dbConn.Exec("UPDATE subscriptions SET service_name=$1, price=$2, user_id=$3, start_date=$4, end_date=$5 WHERE id=$6", sub.NameService, sub.Price, sub.UserId, sub.StartDate, sub.EndDate, id)
	if err != nil{
		http.Error(w, "Не удалось обновить подписку", http.StatusInternalServerError)
		s.Logs.ErrorLog.Println("Ошибка выполнения запроса к бд")
		return
	}

	subsUpdated, err := res.RowsAffected()
	if err != nil{
		http.Error(w, "Не удалось обновить подписку", http.StatusInternalServerError)
		s.Logs.ErrorLog.Println("Не удалось получить кол-во обнвленных записей в бд")
		return
	}

	fmt.Fprintf(w, "В базе обновлена %d подписка\n", subsUpdated)
	
}

//метод получения суммарной стоимости подписок за перид с группировкой по user_id и service_name
func (s *Server) GetSumSubs(w http.ResponseWriter, r *http.Request){
	if r.Method != http.MethodGet{
		http.Error(w, "Недопустимый метод запроса", http.StatusMethodNotAllowed)
		s.Logs.ErrorLog.Println("Неверный метод запроса для подсчета стоимостей подписок")
		return
	}

	// localhost:port/subs/sum?start&end
	//вытаскиваем из url параметры start, end для выбранного периода вида "2025-10-24"	
	start := r.URL.Query().Get("start")
	end := r.URL.Query().Get("end")
	
	//делаем выборку по пользователю и названию подписки, с подсчетом общей стоимости подписок
	rows, err := s.dbConn.Query("SELECT user_id, service_name, sum(price) FROM subscriptions WHERE start_date >= $1 AND end_date <= $2 GROUP BY user_id, service_name", start, end)
	if err != nil{
		http.Error(w, "Не удалось получить информацию", http.StatusInternalServerError)
		s.Logs.ErrorLog.Println("Не удалось выполнить запрос к бд")
		return
	}		
	defer rows.Close()	

	//структура результирующих сумм по сгруппированным и отфильтрованным подпискам
	type RecordSum struct{
		ServiceName string		`json:"service_name"`
		UserId		uuid.UUID	`json:"user_id"`
		Sum			uint		`json:"sum"`
	}

	sums := make([]*RecordSum, 0)	
	//вычитываем записи из бд в слайс результатов		
	for rows.Next(){
		sum := new(RecordSum)

		err := rows.Scan(&sum.UserId, &sum.ServiceName, &sum.Sum)
		if err != nil{
			http.Error(w, "Не удалось получить информацию", http.StatusInternalServerError)
			s.Logs.ErrorLog.Println("Не удалось распарсить записи из бд")
			return
		}
		sums = append(sums, sum)	
	}

	//пишем слайс в тело ответа 
	json.NewEncoder(w).Encode(sums)
}




