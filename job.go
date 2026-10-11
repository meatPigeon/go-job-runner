package main

type JobPayload struct {
	Seed       string `json:"seed"`
	Iterations int    `json:"iterations"`
}
