package domain

import (
	"context"
	transactor "route256/libs/postgres_transactor"
	txMock "route256/libs/postgres_transactor/mocks"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/gojuno/minimock/v3"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
)

func TestCreateOrder(t *testing.T) {
	type repositoryMockFunc func(mc *minimock.Controller) OrdersRepository
	type tmMockFunc func(mc *minimock.Controller) TransactionManager

	type args struct {
		ctx   context.Context
		user  int64
		items []OrderItem
	}

	var (
		mc    = minimock.NewController(t)
		tx    = txMock.NewTxMock(t)
		ctx   = context.Background()
		ctxTx = context.WithValue(ctx, transactor.TxKey("tx"), tx)

		createErr = errors.New("create error")
		updateErr = errors.New("update error")

		orderID = gofakeit.Int64()
		user    = gofakeit.Int64()
		count   = gofakeit.Uint16()
		items   = []OrderItem{
			{
				Sku:   gofakeit.Uint32(),
				Count: count,
			},
			{
				Sku:   gofakeit.Uint32(),
				Count: count,
			},
			{
				Sku:   gofakeit.Uint32(),
				Count: count,
			},
		}
		order = &Order{
			Status: StatusNew,
			User:   user,
			Items:  items,
		}
	)
	t.Cleanup(mc.Finish)

	tests := []struct {
		name           string
		args           args
		want           int64
		err            error
		repositoryMock repositoryMockFunc
		tmMock         tmMockFunc
	}{
		{
			name: "positive case",
			args: args{
				ctx:   ctx,
				user:  user,
				items: items,
			},
			want: orderID,
			err:  nil,
			repositoryMock: func(mc *minimock.Controller) OrdersRepository {
				mock := NewOrdersRepositoryMock(t)
				mock.CreateOrderMock.Expect(ctxTx, order).Return(orderID, nil)
				mock.ReserveStockMock.Set(func(ctx context.Context, orderID int64, item ReservedItem) (err error) {
					return nil
				})
				mock.UpdateOrderStatusMock.Expect(ctxTx, orderID, StatusAwaitingPayment, order.Status).Return(nil)
				mock.StocksMock.Set(func(ctx context.Context, sku uint32) (sa1 []Stock, err error) {
					return []Stock{{
						WarehouseID: gofakeit.Int64(),
						Count:       uint64(count),
					}}, nil
				})
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
			name: "negative case - create order",
			args: args{
				ctx:   ctx,
				user:  user,
				items: items,
			},
			want: 0,
			err:  createErr,
			repositoryMock: func(mc *minimock.Controller) OrdersRepository {
				mock := NewOrdersRepositoryMock(t)
				mock.CreateOrderMock.Expect(ctxTx, order).Return(0, createErr)
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
			name: "negative case - stocks error",
			args: args{
				ctx:   ctx,
				user:  user,
				items: items,
			},
			want: orderID,
			err:  nil,
			repositoryMock: func(mc *minimock.Controller) OrdersRepository {
				mock := NewOrdersRepositoryMock(t)
				mock.CreateOrderMock.Expect(ctxTx, order).Return(orderID, nil)
				mock.ReserveStockMock.Set(func(ctxTx context.Context, orderID int64, item ReservedItem) (err error) {
					return nil
				})
				mock.UpdateOrderStatusMock.Set(func(ctx context.Context, id int64, status string, statusBefore string) (err error) {
					return nil
				})
				mock.StocksMock.Set(func(ctx context.Context, sku uint32) (sa1 []Stock, err error) {
					return nil, errors.New("stocks error")
				})
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
			name: "negative case - no stocks",
			args: args{
				ctx:   ctx,
				user:  user,
				items: items,
			},
			want: orderID,
			err:  nil,
			repositoryMock: func(mc *minimock.Controller) OrdersRepository {
				mock := NewOrdersRepositoryMock(t)
				mock.CreateOrderMock.Expect(ctxTx, order).Return(orderID, nil)
				mock.ReserveStockMock.Set(func(ctxTx context.Context, orderID int64, item ReservedItem) (err error) {
					return nil
				})
				mock.UpdateOrderStatusMock.Set(func(ctx context.Context, id int64, status string, statusBefore string) (err error) {
					return nil
				})
				mock.StocksMock.Set(func(ctx context.Context, sku uint32) (sa1 []Stock, err error) {
					return []Stock{{
						WarehouseID: gofakeit.Int64(),
						Count:       0,
					}}, nil
				})
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
				ctx:   ctx,
				user:  user,
				items: items,
			},
			want: orderID,
			err:  nil,
			repositoryMock: func(mc *minimock.Controller) OrdersRepository {
				mock := NewOrdersRepositoryMock(t)
				mock.CreateOrderMock.Expect(ctxTx, order).Return(orderID, nil)
				mock.ReserveStockMock.Set(func(ctxTx context.Context, orderID int64, item ReservedItem) (err error) {
					return nil
				})
				mock.UpdateOrderStatusMock.Set(func(ctx context.Context, id int64, status string, statusBefore string) (err error) {
					return updateErr
				})
				mock.StocksMock.Set(func(ctx context.Context, sku uint32) (sa1 []Stock, err error) {
					return []Stock{{
						WarehouseID: gofakeit.Int64(),
						Count:       0,
					}}, nil
				})
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
			name: "negative case - update status2",
			args: args{
				ctx:   ctx,
				user:  user,
				items: items,
			},
			want: orderID,
			err:  nil,
			repositoryMock: func(mc *minimock.Controller) OrdersRepository {
				mock := NewOrdersRepositoryMock(t)
				mock.CreateOrderMock.Expect(ctxTx, order).Return(orderID, nil)
				mock.ReserveStockMock.Set(func(ctxTx context.Context, orderID int64, item ReservedItem) (err error) {
					return nil
				})
				mock.UpdateOrderStatusMock.Set(func(ctx context.Context, id int64, status string, statusBefore string) (err error) {
					return updateErr
				})
				mock.StocksMock.Set(func(ctx context.Context, sku uint32) (sa1 []Stock, err error) {
					return []Stock{{
						WarehouseID: gofakeit.Int64(),
						Count:       uint64(count),
					}}, nil
				})
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
			name: "negative case - not full stock",
			args: args{
				ctx:   ctx,
				user:  user,
				items: items,
			},
			want: orderID,
			err:  nil,
			repositoryMock: func(mc *minimock.Controller) OrdersRepository {
				mock := NewOrdersRepositoryMock(t)
				mock.CreateOrderMock.Expect(ctxTx, order).Return(orderID, nil)
				mock.ReserveStockMock.Set(func(ctxTx context.Context, orderID int64, item ReservedItem) (err error) {
					return nil
				})
				mock.UpdateOrderStatusMock.Set(func(ctx context.Context, id int64, status string, statusBefore string) (err error) {
					return nil
				})
				mock.StocksMock.Set(func(ctx context.Context, sku uint32) (sa1 []Stock, err error) {
					return []Stock{{
						WarehouseID: gofakeit.Int64(),
						Count:       uint64(count) + 1,
					}}, nil
				})
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
			name: "negative case - not full stock",
			args: args{
				ctx:   ctx,
				user:  user,
				items: items,
			},
			want: orderID,
			err:  nil,
			repositoryMock: func(mc *minimock.Controller) OrdersRepository {
				mock := NewOrdersRepositoryMock(t)
				mock.CreateOrderMock.Expect(ctxTx, order).Return(orderID, nil)
				mock.ReserveStockMock.Set(func(ctxTx context.Context, orderID int64, item ReservedItem) (err error) {
					return errors.New("reserve err")
				})
				mock.UpdateOrderStatusMock.Set(func(ctx context.Context, id int64, status string, statusBefore string) (err error) {
					return nil
				})
				mock.StocksMock.Set(func(ctx context.Context, sku uint32) (sa1 []Stock, err error) {
					return []Stock{{
						WarehouseID: gofakeit.Int64(),
						Count:       uint64(count),
					}}, nil
				})
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
			api := New(
				tt.repositoryMock(mc),
				tt.tmMock(mc),
			)
			orderID, err := api.CreateOrder(tt.args.ctx, tt.args.user, tt.args.items)
			require.Equal(t, tt.want, orderID)
			if tt.err != nil {
				require.ErrorContains(t, err, tt.err.Error())
			} else {
				require.Equal(t, tt.err, err)
			}
			time.Sleep(3 * time.Second)
		})
	}

}
