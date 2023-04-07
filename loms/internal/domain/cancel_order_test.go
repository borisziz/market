package domain

import (
	"context"
	transactor "route256/libs/postgres_transactor"
	txMock "route256/libs/postgres_transactor/mocks"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
)

func TestCancelOrder(t *testing.T) {
	type repositoryMockFunc func(mc *minimock.Controller) OrdersRepository
	type tmMockFunc func(mc *minimock.Controller) TransactionManager

	type args struct {
		ctx     context.Context
		orderID int64
	}

	var (
		mc    = minimock.NewController(t)
		tx    = txMock.NewTxMock(t)
		ctx   = context.Background()
		ctxTx = context.WithValue(ctx, transactor.TxKey("tx"), tx)

		getOrderErr  = errors.New("get error")
		updateErr    = errors.New("update error")
		unreserveErr = errors.New("unreserve error")

		orderID = gofakeit.Int64()
		order   = &Order{
			ID:     orderID,
			Status: StatusAwaitingPayment,
			User:   gofakeit.Int64(),
			Items:  nil,
		}
		wrongOrder = &Order{
			ID:     orderID,
			Status: StatusFailed,
			User:   gofakeit.Int64(),
			Items:  nil,
		}
	)
	t.Cleanup(mc.Finish)

	tests := []struct {
		name           string
		args           args
		err            error
		repositoryMock repositoryMockFunc
		tmMock         tmMockFunc
	}{
		{
			name: "positive case",
			args: args{
				ctx:     ctx,
				orderID: orderID,
			},
			err: nil,
			repositoryMock: func(mc *minimock.Controller) OrdersRepository {
				mock := NewOrdersRepositoryMock(t)
				mock.GetOrderMock.Expect(ctxTx, orderID).Return(order, nil)
				mock.UpdateOrderStatusMock.Expect(ctxTx, orderID, StatusCancelled, order.Status).Return(nil)
				mock.UnReserveItemsMock.Expect(ctxTx, orderID).Return(nil)
				return mock
			},
			tmMock: func(mc *minimock.Controller) TransactionManager {
				mock := NewTransactionManagerMock(t)
				mock.RunTransactionMock.Set(func(ctx context.Context, isoLevel string, f func(ctxTX context.Context) error) error {
					err := f(ctxTx)
					return err
				})
				return mock
			},
		},
		{
			name: "negative case - get order",
			args: args{
				ctx:     ctx,
				orderID: orderID,
			},
			err: getOrderErr,
			repositoryMock: func(mc *minimock.Controller) OrdersRepository {
				mock := NewOrdersRepositoryMock(t)
				mock.GetOrderMock.Expect(ctxTx, orderID).Return(nil, getOrderErr)
				return mock
			},
			tmMock: func(mc *minimock.Controller) TransactionManager {
				mock := NewTransactionManagerMock(t)
				mock.RunTransactionMock.Set(func(ctx context.Context, isoLevel string, f func(ctxTX context.Context) error) error {
					err := f(ctxTx)
					return err
				})
				return mock
			},
		},
		{
			name: "negative case - update status",
			args: args{
				ctx:     ctx,
				orderID: orderID,
			},
			err: updateErr,
			repositoryMock: func(mc *minimock.Controller) OrdersRepository {
				mock := NewOrdersRepositoryMock(t)
				mock.GetOrderMock.Expect(ctxTx, orderID).Return(order, nil)
				mock.UpdateOrderStatusMock.Expect(ctxTx, orderID, StatusCancelled, order.Status).Return(updateErr)
				return mock
			},
			tmMock: func(mc *minimock.Controller) TransactionManager {
				mock := NewTransactionManagerMock(t)
				mock.RunTransactionMock.Set(func(ctx context.Context, isoLevel string, f func(ctxTX context.Context) error) error {
					err := f(ctxTx)
					return err
				})
				return mock
			},
		},
		{
			name: "negative case - unreserve items",
			args: args{
				ctx:     ctx,
				orderID: orderID,
			},
			err: unreserveErr,
			repositoryMock: func(mc *minimock.Controller) OrdersRepository {
				mock := NewOrdersRepositoryMock(t)
				mock.GetOrderMock.Expect(ctxTx, orderID).Return(order, nil)
				mock.UpdateOrderStatusMock.Expect(ctxTx, orderID, StatusCancelled, order.Status).Return(nil)
				mock.UnReserveItemsMock.Expect(ctxTx, orderID).Return(unreserveErr)
				return mock
			},
			tmMock: func(mc *minimock.Controller) TransactionManager {
				mock := NewTransactionManagerMock(t)
				mock.RunTransactionMock.Set(func(ctx context.Context, isoLevel string, f func(ctxTX context.Context) error) error {
					err := f(ctxTx)
					return err
				})
				return mock
			},
		},
		{
			name: "negative case - wrong status",
			args: args{
				ctx:     ctx,
				orderID: orderID,
			},
			err: errWrongStatus,
			repositoryMock: func(mc *minimock.Controller) OrdersRepository {
				mock := NewOrdersRepositoryMock(t)
				mock.GetOrderMock.Expect(ctxTx, orderID).Return(wrongOrder, nil)
				return mock
			},
			tmMock: func(mc *minimock.Controller) TransactionManager {
				mock := NewTransactionManagerMock(t)
				mock.RunTransactionMock.Set(func(ctx context.Context, isoLevel string, f func(ctxTX context.Context) error) error {
					err := f(ctxTx)
					return err
				})
				return mock
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			api := NewMock(
				tt.repositoryMock(mc),
				tt.tmMock(mc),
			)
			err := api.CancelOrder(tt.args.ctx, tt.args.orderID)
			if tt.err != nil {
				require.ErrorContains(t, err, tt.err.Error())
			} else {
				require.Equal(t, tt.err, err)
			}
		})
	}

}
