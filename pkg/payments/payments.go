package payments

import (
	"context"
	"errors"
	"github.com/DaniilOr/restricted/cmd/service/app/dtos"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v4/pgxpool"
)
var ErrNoToken = errors.New("no token")

type Service struct {
	pool *pgxpool.Pool
}
type Payments struct {
	Id string
	SenderId int64
	Amount int64
}

func NewService(pool *pgxpool.Pool) *Service {
	return &Service{pool: pool}
}

func(s*Service) GetUserPayments(ctx context.Context, id int64)([]*dtos.PaymentDTO, error){
	var payments []*dtos.PaymentDTO
	rows, err := s.pool.Query(ctx,`
	SELECT id, amount FROM payments WHERE senderid=$1 LIMIT 50
	`, id)
	if err != nil{
		return []*dtos.PaymentDTO{}, err
	}
	for rows.Next(){
		var payment dtos.PaymentDTO
		err := rows.Scan(&payment.Id, &payment.Amount)
			if err != nil{
				return []*dtos.PaymentDTO{}, err
			}
		payment.SenderId = id
		payments = append(payments, &payment)
	}
	if rows.Err()!=nil{
		return []*dtos.PaymentDTO{}, rows.Err()
	}
	return payments, nil
}
func(s*Service) AddUserPayments(ctx context.Context, id int64, amount int64)(error){
	uid :=  uuid.New()
	_, err := s.pool.Exec(ctx,`
	INSERT INTO payments(id, senderid, amount) VALUES($1, $2, $3)
	`, uid, id, amount)
	if err != nil{
		return err
	}
	return nil
}
func(s*Service) GetAllPayments(ctx context.Context)([]*dtos.PaymentDTO, error){
	var payments []*dtos.PaymentDTO
	rows, err := s.pool.Query(ctx,`
	SELECT id, senderid, amount FROM payments LIMIT 50
	`,)
	if err != nil{
		return []*dtos.PaymentDTO{}, err
	}
	for rows.Next(){
		var payment dtos.PaymentDTO
		err := rows.Scan(&payment.Id, &payment.SenderId, &payment.Amount)
		if err != nil{
			return []*dtos.PaymentDTO{}, err
		}
		payments = append(payments, &payment)
	}
	if rows.Err()!=nil{
		return []*dtos.PaymentDTO{}, rows.Err()
	}
	return payments, nil
}
