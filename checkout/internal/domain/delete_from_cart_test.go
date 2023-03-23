package domain

import (
	"context"
	transactor "route256/libs/postgres_transactor"
	txMock "route256/libs/postgres_transactor/mocks"
	"testing"

	"github.com/gojuno/minimock/v3"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
)

func TestDeleteFromCart(t *testing.T) {
	type repositoryMockFunc func(mc *minimock.Controller) CartsRepository
	type tmMockFunc func(mc *minimock.Controller) TransactionManager

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

		itemErr   = errors.New("item error")
		deleteErr = errors.New("delete error")

		user       int64  = 1
		sku        uint32 = 4678816
		count      uint16 = 10
		sameCount  uint16 = 15
		largeCount uint16 = 20

		cartItem = &CartItem{
			Sku:   sku,
			Count: 15,
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
				ctx:   ctx,
				user:  user,
				sku:   sku,
				count: count,
			},
			err: nil,
			repositoryMock: func(mc *minimock.Controller) CartsRepository {
				mock := NewCartsRepositoryMock(t)
				mock.GetCartItemMock.Expect(ctxTx, user, sku).Return(cartItem, nil)
				mock.DeleteFromCartMock.Expect(ctxTx, user, sku, count, cartItem.Count == count).Return(nil)
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
			name: "negative case - item error",
			args: args{
				ctx:   ctx,
				user:  user,
				sku:   sku,
				count: count,
			},
			err: itemErr,
			repositoryMock: func(mc *minimock.Controller) CartsRepository {
				mock := NewCartsRepositoryMock(t)
				mock.GetCartItemMock.Expect(ctxTx, user, sku).Return(nil, itemErr)
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
			name: "negative case - delete error",
			args: args{
				ctx:   ctx,
				user:  user,
				sku:   sku,
				count: count,
			},
			err: deleteErr,
			repositoryMock: func(mc *minimock.Controller) CartsRepository {
				mock := NewCartsRepositoryMock(t)
				mock.GetCartItemMock.Expect(ctxTx, user, sku).Return(cartItem, nil)
				mock.DeleteFromCartMock.Expect(ctxTx, user, sku, count, count == cartItem.Count).Return(deleteErr)
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
			name: "negative case - large count",
			args: args{
				ctx:   ctx,
				user:  user,
				sku:   sku,
				count: largeCount,
			},
			err: ErrNoSoManyItems,
			repositoryMock: func(mc *minimock.Controller) CartsRepository {
				mock := NewCartsRepositoryMock(t)
				mock.GetCartItemMock.Expect(ctxTx, user, sku).Return(cartItem, nil)
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
			name: "positive with same count",
			args: args{
				ctx:   ctx,
				user:  user,
				sku:   sku,
				count: sameCount,
			},
			err: nil,
			repositoryMock: func(mc *minimock.Controller) CartsRepository {
				mock := NewCartsRepositoryMock(t)
				mock.GetCartItemMock.Expect(ctxTx, user, sku).Return(cartItem, nil)
				mock.DeleteFromCartMock.Expect(ctxTx, user, sku, sameCount, sameCount == cartItem.Count).Return(nil)
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
				tt.repositoryMock(mc),
				tt.tmMock(mc),
			)
			if err != nil {
				require.Equal(t, nil, err)
			}
			err = api.DeleteFromCart(tt.args.ctx, tt.args.user, tt.args.sku, tt.args.count)
			if tt.err != nil {
				require.ErrorContains(t, err, tt.err.Error())
			} else {
				require.Equal(t, tt.err, err)
			}
		})
	}

}
