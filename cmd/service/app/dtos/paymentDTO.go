package dtos

type PaymentDTO struct{
	Id string `json:"id"`
	SenderId int64 `json:"sender_id"`
	Amount int64 `json:"amount"`
}
