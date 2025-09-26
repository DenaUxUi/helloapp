package main

import (
    "encoding/json"
    "net/http"
    "fmt"
    "log"
)
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
    mux := http.NewServeMux()
    mux.HandleFunc("/api/hello", formHandler)

    http.ListenAndServe(":8000", corsMiddleware(mux))
    fmt.Println("Starting server at port 8000")
}