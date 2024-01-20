package data

import "time"

type Payment struct {
	Id            int       `json:"id"`
	ClientId      int       `json:"clientId"`
	Value         int       `json:"value"`
	OperationDate time.Time `json:"operationDate"`
}
