package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

type Response struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Lang    string `json:"language"`
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		res := Response{
			Status:  "success",
			Message: "Halo dari Go Binary di dalam container SCRATCH (Kosong)!",
			Lang:    "Go (Golang)",
		}
		json.NewEncoder(w).Encode(res)
	})

	fmt.Printf("Server Go berjalan di port %s...\n", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		fmt.Printf("Gagal menjalankan server: %v\n", err)
	}
}
