package domain

import (
	"context"
	transactor "route256/libs/postgres_transactor"
	txMock "route256/libs/postgres_transactor/mocks"
	"testing"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/gojuno/minimock/v3"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
)

func TestAddToCart(t *testing.T) {
	type repositoryMockFunc func(mc *minimock.Controller) CartsRepository
	type productsMockFunc func(mc *minimock.Controller) ProductServiceCaller
	type tmMockFunc func(mc *minimock.Controller) TransactionManager
	type lomsMockFunc func(mc *minimock.Controller) LOMSCaller

	type args struct {
		ctx   context.Context
		user  int64
		sku   uint32
		count uint16
	}

	var (
		mc    = minimock.NewController(t)
		tx    = txMock.NewTxMock(t)
		ctx   = context.Background()
		ctxTx = context.WithValue(ctx, transactor.TxKey("tx"), tx)

		itemErr      = errors.New("item error")
		stocksErr    = errors.New("stocks error")
		addToCartErr = errors.New("add error")

		user       int64  = 1
		sku        uint32 = 4678816
		count      uint16 = 10
		largeCount uint16 = 30

		skus = map[uint32]struct{}{
			4678816: {},
		}
		cartItem = &CartItem{
			Sku:   sku,
			Count: 15,
		}
		invalidSku uint32 = 1

		stocks = []Stock{
			{
				WarehouseID: gofakeit.Int64(),
				Count:       10,
			},
			{
				WarehouseID: gofakeit.Int64(),
				Count:       10,
			},
			{
				WarehouseID: gofakeit.Int64(),
				Count:       5,
			},
		}
	)
	t.Cleanup(mc.Finish)

	tests := []struct {
		name           string
		args           args
		err            error
		repositoryMock repositoryMockFunc
		productsMock   productsMockFunc
		lomsMock       lomsMockFunc
		tmMock         tmMockFunc
	}{
		{
			name: "positive case",
			args: args{
				ctx:   ctx,
				user:  user,
				sku:   sku,
				count: count,
			},
			err: nil,
			productsMock: func(mc *minimock.Controller) ProductServiceCaller {
				mock := NewProductServiceCallerMock(t)
				mock.GetSKUsMock.Expect(ctx).Return(skus, nil)
				return mock
			},
			repositoryMock: func(mc *minimock.Controller) CartsRepository {
				mock := NewCartsRepositoryMock(t)
				mock.GetCartItemMock.Expect(ctxTx, user, sku).Return(cartItem, nil)
				mock.AddToCartMock.Expect(ctxTx, user, sku, count).Return(nil)
				return mock
			},
			lomsMock: func(mc *minimock.Controller) LOMSCaller {
				mock := NewLOMSCallerMock(t)
				mock.StocksMock.Expect(ctxTx, sku).Return(stocks, nil)
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
			name: "negative case - products error",
			args: args{
				ctx:   ctx,
				user:  user,
				sku:   invalidSku,
				count: count,
			},
			err: ErrInvalidSKU,
			productsMock: func(mc *minimock.Controller) ProductServiceCaller {
				mock := NewProductServiceCallerMock(t)
				mock.GetSKUsMock.Expect(ctx).Return(skus, nil)
				return mock
			},
			repositoryMock: func(mc *minimock.Controller) CartsRepository {
				mock := NewCartsRepositoryMock(t)
				return mock
			},
			lomsMock: func(mc *minimock.Controller) LOMSCaller {
				mock := NewLOMSCallerMock(t)
				return mock
			},
			tmMock: func(mc *minimock.Controller) TransactionManager {
				mock := NewTransactionManagerMock(t)
				return mock
			},
		},
		{
			name: "negative case - item error",
			args: args{
				ctx:   ctx,
				user:  user,
				sku:   sku,
				count: count,
			},
			err: itemErr,
			productsMock: func(mc *minimock.Controller) ProductServiceCaller {
				mock := NewProductServiceCallerMock(t)
				mock.GetSKUsMock.Expect(ctx).Return(skus, nil)
				return mock
			},
			repositoryMock: func(mc *minimock.Controller) CartsRepository {
				mock := NewCartsRepositoryMock(t)
				mock.GetCartItemMock.Expect(ctxTx, user, sku).Return(nil, itemErr)
				return mock
			},
			lomsMock: func(mc *minimock.Controller) LOMSCaller {
				mock := NewLOMSCallerMock(t)
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
				sku:   sku,
				count: count,
			},
			err: stocksErr,
			productsMock: func(mc *minimock.Controller) ProductServiceCaller {
				mock := NewProductServiceCallerMock(t)
				mock.GetSKUsMock.Expect(ctx).Return(skus, nil)
				return mock
			},
			repositoryMock: func(mc *minimock.Controller) CartsRepository {
				mock := NewCartsRepositoryMock(t)
				mock.GetCartItemMock.Expect(ctxTx, user, sku).Return(cartItem, nil)
				return mock
			},
			lomsMock: func(mc *minimock.Controller) LOMSCaller {
				mock := NewLOMSCallerMock(t)
				mock.StocksMock.Expect(ctxTx, sku).Return(nil, stocksErr)
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
			name: "negative case - insufficient stocks",
			args: args{
				ctx:   ctx,
				user:  user,
				sku:   sku,
				count: largeCount,
			},
			err: ErrInsufficientStocks,
			productsMock: func(mc *minimock.Controller) ProductServiceCaller {
				mock := NewProductServiceCallerMock(t)
				mock.GetSKUsMock.Expect(ctx).Return(skus, nil)
				return mock
			},
			repositoryMock: func(mc *minimock.Controller) CartsRepository {
				mock := NewCartsRepositoryMock(t)
				mock.GetCartItemMock.Expect(ctxTx, user, sku).Return(cartItem, nil)
				return mock
			},
			lomsMock: func(mc *minimock.Controller) LOMSCaller {
				mock := NewLOMSCallerMock(t)
				mock.StocksMock.Expect(ctxTx, sku).Return(stocks, nil)
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
			name: "negative case - add error",
			args: args{
				ctx:   ctx,
				user:  user,
				sku:   sku,
				count: count,
			},
			err: addToCartErr,
			productsMock: func(mc *minimock.Controller) ProductServiceCaller {
				mock := NewProductServiceCallerMock(t)
				mock.GetSKUsMock.Expect(ctx).Return(skus, nil)
				return mock
			},
			repositoryMock: func(mc *minimock.Controller) CartsRepository {
				mock := NewCartsRepositoryMock(t)
				mock.GetCartItemMock.Expect(ctxTx, user, sku).Return(cartItem, nil)
				mock.AddToCartMock.Expect(ctxTx, user, sku, count).Return(addToCartErr)
				return mock
			},
			lomsMock: func(mc *minimock.Controller) LOMSCaller {
				mock := NewLOMSCallerMock(t)
				mock.StocksMock.Expect(ctxTx, sku).Return(stocks, nil)
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
			name: "positive with empty cart",
			args: args{
				ctx:   ctx,
				user:  user,
				sku:   sku,
				count: count,
			},
			err: nil,
			productsMock: func(mc *minimock.Controller) ProductServiceCaller {
				mock := NewProductServiceCallerMock(t)
				mock.GetSKUsMock.Expect(ctx).Return(skus, nil)
				return mock
			},
			repositoryMock: func(mc *minimock.Controller) CartsRepository {
				mock := NewCartsRepositoryMock(t)
				mock.GetCartItemMock.Expect(ctxTx, user, sku).Return(nil, ErrNoSameItemsInCart)
				mock.AddToCartMock.Expect(ctxTx, user, sku, count).Return(nil)
				return mock
			},
			lomsMock: func(mc *minimock.Controller) LOMSCaller {
				mock := NewLOMSCallerMock(t)
				mock.StocksMock.Expect(ctxTx, sku).Return(stocks, nil)
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
			api, err := NewMock(
				tt.productsMock(mc),
				tt.repositoryMock(mc),
				tt.tmMock(mc),
				tt.lomsMock(mc),
			)
			if err != nil {
				require.Equal(t, nil, err)
			}
			err = api.AddToCart(tt.args.ctx, tt.args.user, tt.args.sku, tt.args.count)
			if tt.err != nil {
				require.ErrorContains(t, err, tt.err.Error())
			} else {
				require.Equal(t, tt.err, err)
			}
		})
	}

}
