package models

type AxiomPayload struct {
	Time       string `json:"_time"`
	Throughput uint64 `json:"throughput"`
}
