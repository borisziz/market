package domain

import (
	"context"
	"testing"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/gojuno/minimock/v3"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
)

func TestListOrder(t *testing.T) {
	type repositoryMockFunc func(mc *minimock.Controller) OrdersRepository
	type tmMockFunc func(mc *minimock.Controller) TransactionManager

	type args struct {
		ctx     context.Context
		orderID int64
	}

	var (
		mc  = minimock.NewController(t)
		ctx = context.Background()

		getOrderErr = errors.New("get error")

		orderID = gofakeit.Int64()
		order   = &Order{
			ID:     orderID,
			Status: StatusAwaitingPayment,
			User:   gofakeit.Int64(),
			Items:  nil,
		}
	)
	t.Cleanup(mc.Finish)

	tests := []struct {
		name           string
		args           args
		want           *Order
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
			want: order,
			err:  nil,
			repositoryMock: func(mc *minimock.Controller) OrdersRepository {
				mock := NewOrdersRepositoryMock(t)
				mock.GetOrderMock.Expect(ctx, orderID).Return(order, nil)
				return mock
			},
			tmMock: func(mc *minimock.Controller) TransactionManager {
				mock := NewTransactionManagerMock(t)
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
				mock.GetOrderMock.Expect(ctx, orderID).Return(nil, getOrderErr)
				return mock
			},
			tmMock: func(mc *minimock.Controller) TransactionManager {
				mock := NewTransactionManagerMock(t)
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
			order, err := api.ListOrder(tt.args.ctx, tt.args.orderID)
			require.Equal(t, tt.want, order)
			if tt.err != nil {
				require.ErrorContains(t, err, tt.err.Error())
			} else {
				require.Equal(t, tt.err, err)
			}
		})
	}

}
