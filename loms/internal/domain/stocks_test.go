package domain

import (
	"context"
	"testing"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/gojuno/minimock/v3"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
)

func TestStocks(t *testing.T) {
	type repositoryMockFunc func(mc *minimock.Controller) OrdersRepository
	type tmMockFunc func(mc *minimock.Controller) TransactionManager

	type args struct {
		ctx context.Context
		sku uint32
	}

	var (
		mc  = minimock.NewController(t)
		ctx = context.Background()

		stocksErr = errors.New("stocks error")
		sku       = gofakeit.Uint32()
		stocks    = []Stock{
			{
				WarehouseID: gofakeit.Int64(),
				Count:       gofakeit.Uint64(),
			},
			{
				WarehouseID: gofakeit.Int64(),
				Count:       gofakeit.Uint64(),
			},
			{
				WarehouseID: gofakeit.Int64(),
				Count:       gofakeit.Uint64(),
			},
		}
	)
	t.Cleanup(mc.Finish)

	tests := []struct {
		name           string
		args           args
		want           []Stock
		err            error
		repositoryMock repositoryMockFunc
		tmMock         tmMockFunc
	}{
		{
			name: "positive case",
			args: args{
				ctx: ctx,
				sku: sku,
			},
			want: stocks,
			err:  nil,
			repositoryMock: func(mc *minimock.Controller) OrdersRepository {
				mock := NewOrdersRepositoryMock(t)
				mock.StocksMock.Expect(ctx, sku).Return(stocks, nil)
				return mock
			},
			tmMock: func(mc *minimock.Controller) TransactionManager {
				mock := NewTransactionManagerMock(t)
				return mock
			},
		},
		{
			name: "negative case - stocks error",
			args: args{
				ctx: ctx,
				sku: sku,
			},
			want: nil,
			err:  stocksErr,
			repositoryMock: func(mc *minimock.Controller) OrdersRepository {
				mock := NewOrdersRepositoryMock(t)
				mock.StocksMock.Expect(ctx, sku).Return(nil, stocksErr)
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
			result, err := api.Stocks(tt.args.ctx, tt.args.sku)
			require.Equal(t, tt.want, result)
			if tt.err != nil {
				require.ErrorContains(t, err, tt.err.Error())
			} else {
				require.Equal(t, tt.err, err)
			}
		})
	}

}
