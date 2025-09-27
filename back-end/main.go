package main

import (
    "database/sql"
    "encoding/json"
    "net/http"
    "fmt"
    "math/rand"
    "log"
    "os"
    "time"
    "strconv"

    "github.com/gorilla/mux"
    _ "github.com/lib/pq"
    // "github.com/joho/godotenv"
)

var db *sql.DB


var (
    host     string
    port     string
    user     string
    password string
    dbname   string
)



func init() {
    host = os.Getenv("POSTGRES_HOST")
    port = os.Getenv("POSTGRES_PORT")
    user = os.Getenv("POSTGRES_USER")
    password = os.Getenv("POSTGRES_PASSWORD")
    dbname = os.Getenv("POSTGRES_DB")
}

// структура для входящего запроса
type RequestData struct {
    Name string `json:"name"`
}

// структура для ответа
type ResponseData struct {
    Message string `json:"message"`
}

func formHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        http.Error(w, "Only POST allowed", http.StatusMethodNotAllowed)
        return
    }

    w.Header().Set("Content-Type", "application/json")

    var req RequestData
    err := json.NewDecoder(r.Body).Decode(&req)
    if err != nil {
        http.Error(w, `{"message":"Invalid JSON"}`, http.StatusBadRequest)
        return
    }

    // формируем ответ
    resp := ResponseData{
        Message: "Hello, " + req.Name + "!",
    }

    // отправляем JSON
    if err := json.NewEncoder(w).Encode(resp); err != nil {
        log.Println("Failed to write JSON:", err)
    }
}
func createInstance(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodGet {
        http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
        return
    }

    rand.Seed(time.Now().UnixNano()) // инициализация генератора случайных чисел
    instanceId := rand.Intn(900000) + 100000 // 6-значный ID

    // Вставка в БД с плейсхолдером
    _, err := db.Exec("INSERT INTO instances (instance) VALUES ($1)", instanceId)
    if err != nil {
        http.Error(w, "Failed to insert into DB: "+err.Error(), http.StatusInternalServerError)
        log.Printf("Created instance with ID %d\n", instanceId)

        return
    }

    // Отправляем ответ
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(map[string]interface{}{
        "instance_id": instanceId,
        "status":      "created",
    })
}
func terminateInstance(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    idStr := vars["id"]

    id, err := strconv.Atoi(idStr)
    if err != nil {
        http.Error(w, "Invalid id", http.StatusBadRequest)
        return
    }

    // Выполняем удаление из базы
    result, err := db.Exec("DELETE FROM instances WHERE id = $1", id)
    if err != nil {
        http.Error(w, "Failed to delete instance: "+err.Error(), http.StatusInternalServerError)
        return
    }

    rowsAffected, err := result.RowsAffected()
    if err != nil {
        http.Error(w, "Error checking deletion result: "+err.Error(), http.StatusInternalServerError)
        return
    }

    if rowsAffected == 0 {
        http.Error(w, "Instance not found", http.StatusNotFound)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    w.Write([]byte(fmt.Sprintf(`{"status":"success","deleted_id":%d}`, id)))
}
func listInstance(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodGet {
        http.Error(w, "Only GET method is allowed", http.StatusMethodNotAllowed)
        return
    }

    rows, err := db.Query("SELECT instance FROM instances;")
    if err != nil {
        http.Error(w, "Failed to fetch instances: "+err.Error(), http.StatusInternalServerError)
        return
    }
    defer rows.Close()

    var instances []int
    for rows.Next() {
        var instanceId int
        if err := rows.Scan(&instanceId); err != nil {
            http.Error(w, "Failed to scan row: "+err.Error(), http.StatusInternalServerError)
            return
        }
        instances = append(instances, instanceId)
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]interface{}{
        "instances": instances,
    })
}




func corsMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Разрешаем любые источники (для dev), можно ограничить позже
        w.Header().Set("Access-Control-Allow-Origin", "*")
        w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
        w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

        // Preflight-запрос
        if r.Method == http.MethodOptions {
            w.WriteHeader(http.StatusOK)
            return
        }

        next.ServeHTTP(w, r)
    })
}

func main() {
    psqlInfo := fmt.Sprintf(
        "host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
        host, port, user, password, dbname,
    )


    fmt.Println("Connecting to DB:", host, port, user, dbname)

    var err error    
    db, err = sql.Open("postgres", psqlInfo)
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close() 

     err = db.Ping()
    if err != nil {
        log.Fatal("Не удалось подключиться к БД:", err)
    }

    fmt.Println("Успешное подключение к базе данных!")
    r := mux.NewRouter()

    // обработчики
    r.HandleFunc("/api/hello", formHandler).Methods("POST")
    r.HandleFunc("/api/ec2/create", createInstance).Methods("GET")
    r.HandleFunc("/api/ec2/list", listInstance).Methods("GET")
    r.HandleFunc("/api/ec/terminate/{id}", terminateInstance).Methods("POST")
    
    // CORS
    handler := corsMiddleware(r)
    
    fmt.Println("Starting server at port 8000")
    http.ListenAndServe(":8000", handler)

}