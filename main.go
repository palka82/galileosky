package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"sync"
	"time"
)

const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
var ID = 0

//Список устройств
type DevList struct {
	Name string `json:"name"`//Имя устройства
	Id int `json:"id"`// ID устройства
	DT string `json:"date"`//Дата и время
}

//Структура ответа
type Answer struct {
	Type    string `json:"type"` //Тип - "ОК" || "ERR"
	Message string `json:"message"` //сообщение
}

//Ответ для устройств
type AnswerDevList struct {
	Type string `json:"type"`
	Message []DevList `json:"message"`
}
//Ответ для устройств со статистикой
type AnswerDevListStats struct {
	Type string `json:"type"`
	Message []Event `json:"message"`
}


type DevCommand struct {
	//Структура для команд
	Command string `json:"command"`
	Params  struct {
		DT string `json:"dt"`
		Name string `json:"name"`
		User string `json:"user"`
		Temp float32 `json:"temp"`
	} `json:"params"`
}

type UserCommand struct {
	//Структура для команд
	Command string `json:"command"`
	Params  string `json:"params"`
}

type Event struct {
	DT string //Дата и время показаний
	devName string //Имя устройства
	Temp float32 //Температура
}

type Device struct {
	Name string
	User string
	Token  string
	Id int
}


var DevicesList = []Device{} //Список устройств
var EventsList = []Event{} //Список событий
var mu sync.Mutex

var Users = make(map[string]string)              //Список пользователей

var seededRand *rand.Rand = rand.New(rand.NewSource(time.Now().UnixNano()))

//Функция для формирования строки из символов
func randSeq(length byte, chars string) string {
	byt := make([]byte, length)
	for i := range byt {
		byt[i] = chars[seededRand.Intn(len(chars))]
	}
	return string(byt)
}

//Функция для генерации токенов
func generateToken(length byte) string {
	return randSeq(length, chars)
}

func main() {
	http.HandleFunc("/", handler)
	http.HandleFunc("/users", handlerUser)
	http.HandleFunc("/devices", handlerDevices)
	http.HandleFunc("/stats", handlerStats)
	log.Fatal(http.ListenAndServe("localhost:8000", nil))
}

func handler(w http.ResponseWriter, r *http.Request) {
	log.Printf("%s %s %s\n", r.Method, r.URL, r.Proto)
	w.WriteHeader(http.StatusNotFound)
}

func handlerUser(w http.ResponseWriter, r *http.Request) {
	log.Printf("%s %s %s\n", r.Method, r.URL, r.Proto)
	r.ParseForm() //парсим аргументы
	header := w.Header()
	//Метод Post
	if r.Method == "POST" {
		//Проверям наличие данных
		if r.Form["data"] == nil {
			header.Set("Content-Type", "application/json")
			response := Answer{
				Type:    "ERR",
				Message: "Variable \"data\" not set!",
			}
			data, err := json.Marshal(response)
			if err != nil {
				log.Printf("%s", err)
				w.WriteHeader(http.StatusInternalServerError)
				fmt.Fprintf(w, "%s", err)
				return
			}
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, "%s", data)
		} else {
			//Данные есть - парсим
			var dat UserCommand

			if err := json.Unmarshal([]byte(r.Form["data"][0]), &dat); err != nil {
				log.Printf("%s", err)
				w.WriteHeader(http.StatusInternalServerError)
				fmt.Fprintf(w, "%s", err)
				return
			}
			//Регистрация нового пользователя
			if dat.Command == "reg" {
				token := generateToken(20)
				mu.Lock()
				Users[dat.Params] = token
				mu.Unlock()
				header.Set("Content-Type", "application/json")
				response := Answer{
					Type:    "OK",
					Message: token,
				}
				data, err := json.Marshal(response)
				if err != nil {
					log.Printf("%s", err)
					w.WriteHeader(http.StatusInternalServerError)
					fmt.Fprintf(w, "%s", err)
					return
				}
				w.WriteHeader(http.StatusOK)
				fmt.Fprintf(w, "%s", data)
			}
		}
	}
}

func handlerDevices(w http.ResponseWriter, r *http.Request) {
	log.Printf("%s %s %s\n", r.Method, r.URL, r.Proto)
	r.ParseForm() //парсим аргументы
	header := w.Header()
	if r.Method == "POST" {
		//Проверям наличие данных
		if r.Form["data"] == nil {
			header.Set("Content-Type", "application/json")
			response := Answer{
				Type:    "ERR",
				Message: "Variable \"data\" not set!",
			}
			data, err := json.Marshal(response)
			if err != nil {
				log.Printf("%s", err)
				w.WriteHeader(http.StatusInternalServerError)
				fmt.Fprintf(w, "%s", err)
				return
			}
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, "%s", data)
		} else {
			//Данные есть - парсим
			var dat DevCommand

			if err := json.Unmarshal([]byte(r.Form["data"][0]), &dat); err != nil {
				log.Printf("%s", err)
				w.WriteHeader(http.StatusInternalServerError)
				fmt.Fprintf(w, "%s", err)
				return
			}

			//Регистрация нового устройства
			if dat.Command == "reg" {
				if Users[dat.Params.User] == "" {
						header.Set("Content-Type", "application/json")
						response := Answer{
							Type:    "ERR",
							Message: "User not found!",
						}
						data, err := json.Marshal(response)
						if err != nil {
							log.Printf("%s", err)
							w.WriteHeader(http.StatusInternalServerError)
							fmt.Fprintf(w, "%s", err)
							return
						}
						w.WriteHeader(http.StatusOK)
						fmt.Fprintf(w, "%s", data)
						return
				}
				token := generateToken(20)
				ID++
				device := Device{
					Name: dat.Params.Name,
					User: dat.Params.User,
					Token: token,
					Id: ID,
				}
				mu.Lock()
				DevicesList = append(DevicesList, device)
				mu.Unlock()
				header.Set("Content-Type", "application/json")
				response := Answer{
					Type:    "OK",
					Message: token,
				}
				data, err := json.Marshal(response)
				if err != nil {
					log.Printf("%s", err)
					w.WriteHeader(http.StatusInternalServerError)
					fmt.Fprintf(w, "%s", err)
					return
				}
				w.WriteHeader(http.StatusOK)
				fmt.Fprintf(w, "%s", data)
			}
			//Сохранение данных
			if dat.Command == "put" {
				auth := r.Header.Get("Authorization")
				if auth == "" {
					header.Set("Content-Type", "application/json")
					response := Answer{
						Type:    "ERR",
						Message: "Token not found!",
					}
					data, err := json.Marshal(response)
					if err != nil {
						log.Printf("%s", err)
						w.WriteHeader(http.StatusInternalServerError)
						fmt.Fprintf(w, "%s", err)
						return
					}
					w.WriteHeader(http.StatusUnauthorized)
					fmt.Fprintf(w, "%s", data)
					return
				}
				for _, dev := range DevicesList {
					if dev.Token == auth {
						event := Event{
							DT: dat.Params.DT,
							devName: dev.Name,
							Temp: dat.Params.Temp,
						}
						mu.Lock()
						EventsList = append(EventsList, event)
						mu.Unlock()
						w.WriteHeader(http.StatusCreated)
						return
					}
				}
				//Не нашелся токен
				header.Set("Content-Type", "application/json")
				response := Answer{
					Type:    "ERR",
					Message: "Token not found!",
				}
				data, err := json.Marshal(response)
				if err != nil {
					log.Printf("%s", err)
					w.WriteHeader(http.StatusInternalServerError)
					fmt.Fprintf(w, "%s", err)
					return
				}
				w.WriteHeader(http.StatusUnauthorized)
				fmt.Fprintf(w, "%s", data)
				return
			}
		}
	}
	if r.Method == "GET" {
		auth := r.Header.Get("Authorization")
		devList := []DevList{}
		if auth == "" {
			header.Set("Content-Type", "application/json")
			response := Answer{
				Type:    "ERR",
				Message: "Token not found!",
			}
			data, err := json.Marshal(response)
			if err != nil {
				log.Printf("%s", err)
				w.WriteHeader(http.StatusInternalServerError)
				fmt.Fprintf(w, "%s", err)
				return
			}
			w.WriteHeader(http.StatusUnauthorized)
			fmt.Fprintf(w, "%s", data)
			return
		}
		for user := range Users {
			//Нашелся токен для пользователя
			if Users[user] == auth {
				//Перебор устройств
				for _, dev := range DevicesList {
					//Нашлось устройство
					if dev.User == user {
						dt := ""
						for _, e := range EventsList {
							if e.devName == dev.Name {
								dt = e.DT
							}
						}
						d := DevList{Name: dev.Name, Id: dev.Id, DT: dt}
						mu.Lock()
						devList = append(devList, d)
						mu.Unlock()
					}
				}

				response := &AnswerDevList{
					Type: "OK",
					Message: devList,
				}
				data, err := json.Marshal(response)
				if err != nil {
					log.Printf("%s", err)
					w.WriteHeader(http.StatusInternalServerError)
					fmt.Fprintf(w, "%s", err)
					return
				}
				w.WriteHeader(http.StatusOK)
				fmt.Fprintf(w, "%s", data)
				return
			}
		}

		for _, dev := range DevicesList {
			if dev.Token == auth {
				for _, e := range EventsList {
					if e.devName == dev.Name {

					}
				}
			}
		}
		//Не нашли
		header.Set("Content-Type", "application/json")
		response := Answer{
			Type:    "ERR",
			Message: "Token not found!",
		}
		data, err := json.Marshal(response)
		if err != nil {
			log.Printf("%s", err)
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "%s", err)
			return
		}
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprintf(w, "%s", data)
		return
	}
}

func handlerStats(w http.ResponseWriter, r *http.Request) {
	log.Printf("%s %s %s\n", r.Method, r.URL, r.Proto)
	r.ParseForm() //парсим аргументы
	header := w.Header()
	eventList := []Event{}
	if r.Method == "GET" {
		auth := r.Header.Get("Authorization")
		if auth == "" {
			header.Set("Content-Type", "application/json")
			response := Answer{
				Type:    "ERR",
				Message: "Token not found!",
			}
			data, err := json.Marshal(response)
			if err != nil {
				log.Printf("%s", err)
				w.WriteHeader(http.StatusInternalServerError)
				fmt.Fprintf(w, "%s", err)
				return
			}
			w.WriteHeader(http.StatusUnauthorized)
			fmt.Fprintf(w, "%s", data)
			return
		}

		ids, ok := r.URL.Query()["id"]
		if !ok || len(ids[0]) < 1 {
			header.Set("Content-Type", "application/json")
			response := Answer{
				Type:    "ERR",
				Message: "ID not set!",
			}
			data, err := json.Marshal(response)
			if err != nil {
				log.Printf("%s", err)
				w.WriteHeader(http.StatusInternalServerError)
				fmt.Fprintf(w, "%s", err)
				return
			}
			w.WriteHeader(http.StatusUnauthorized)
			fmt.Fprintf(w, "%s", data)
			return
		}
		id, err := strconv.Atoi(ids[0])
		if err != nil {
			response := Answer{
				Type:    "ERR",
				Message: "Atoi error",
			}
			data, err := json.Marshal(response)
			if err != nil {
				log.Printf("%s", err)
				w.WriteHeader(http.StatusInternalServerError)
				fmt.Fprintf(w, "%s", err)
				return
			}
			w.WriteHeader(http.StatusUnauthorized)
			fmt.Fprintf(w, "%s", data)
			return
		}


		for user := range Users {
			if Users[user] == auth {
				//Перебор устройств
				for _, dev := range DevicesList {
					//Нашлось устройство
					if dev.User == user && dev.Id == id {
						for _, e := range EventsList {
							if e.devName == dev.Name {
								mu.Lock()
								eventList = append(eventList, e)
								mu.Unlock()
							}
						}
					}
				}
				response := &AnswerDevListStats{
					Type: "OK",
					Message: eventList,
				}
				data, err := json.Marshal(response)
				if err != nil {
					log.Printf("%s", err)
					w.WriteHeader(http.StatusInternalServerError)
					fmt.Fprintf(w, "%s", err)
					return
				}
				w.WriteHeader(http.StatusOK)
				fmt.Fprintf(w, "%s", data)
				return
			}
		}
	}
	//Не нашли
	header.Set("Content-Type", "application/json")
	response := Answer{
		Type:    "ERR",
		Message: "Token not found!",
	}
	data, err := json.Marshal(response)
	if err != nil {
		log.Printf("%s", err)
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "%s", err)
		return
	}
	w.WriteHeader(http.StatusUnauthorized)
	fmt.Fprintf(w, "%s", data)
	return
}