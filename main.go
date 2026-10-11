package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
)

func main() {
	fmt.Println("job runner starting")
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "OK")
	})

	db, err := openDatabase()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	fmt.Println("database connected")

	http.HandleFunc("/jobs", func(w http.ResponseWriter, r *http.Request) {

		if r.Method != http.MethodPost {
			w.Header().Set("Allow", http.MethodPost)
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var payload JobPayload

		err := json.NewDecoder(r.Body).Decode(&payload)
		if err != nil {
			http.Error(w, "invalid JSON", http.StatusBadRequest)
			return
		}

		if payload.Iterations < 1 || payload.Iterations > 100000000 {
			http.Error(w, "iterations must be between 1 and 100 000 000", http.StatusBadRequest)
			return
		}

		if len(payload.Seed) < 1 || len(payload.Seed) > 256 {
			http.Error(w, "seed length must be between 1 and 256", http.StatusBadRequest)
			return
		}

		payloadJSON, err := json.Marshal(payload)

		if err != nil {
			http.Error(w, "invalid payload", http.StatusInternalServerError)
			return
		}

		var id int64
		err = db.QueryRowContext(
			r.Context(),
			"INSERT INTO jobs (payload) VALUES ($1) RETURNING id",
			string(payloadJSON),
		).Scan(&id)

		if err != nil {
			http.Error(w, "database unavailable", http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusAccepted)
		fmt.Fprintf(w, "id=%d\n", id)
	})
	http.HandleFunc("GET /jobs/{id}", func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
		if err != nil || id < 1 {
			http.Error(w, "invalid job ID", http.StatusBadRequest)
			return
		}

		var state string
		err = db.QueryRowContext(
			r.Context(),
			"SELECT state FROM jobs WHERE id = $1",
			id,
		).Scan(&state)
		if err == sql.ErrNoRows {
			http.Error(w, "job not found", http.StatusNotFound)
			return
		}
		if err != nil {
			http.Error(w, "database unavailable", http.StatusServiceUnavailable)
			return
		}

		// TODO: Respond with the job ID and state using fmt.Fprintf.
		fmt.Fprintf(w, "id=%d, state=%s\n", id, state)
	})

	log.Fatal(http.ListenAndServe("127.0.0.1:8080", nil))
}
