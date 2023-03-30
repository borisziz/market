package domain

import (
	"context"
	"testing"

	"github.com/gojuno/minimock/v3"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
)

func TestPurchase(t *testing.T) {
	type repositoryMockFunc func(mc *minimock.Controller) CartsRepository
	type lomsCallerMockFunc func(mc *minimock.Controller) LOMSCaller

	type args struct {
		ctx  context.Context
		user int64
	}

	var (
		mc                 = minimock.NewController(t)
		ctx                = context.Background()
		lomsRes      int64 = 5
		lomsErrorRes int64 = 0
		orderIDError       = lomsErrorRes
		repoErr            = errors.New("repo error")
		lomsErr            = errors.New("loms error")
		user         int64 = 1

		cartItems = []CartItem{
			{
				Sku:   1148162,
				Count: 1,
			},
			{
				Sku:   6967749,
				Count: 2,
			},
		}
		emptyCart = make([]CartItem, 0)
		orderID   = lomsRes
	)
	t.Cleanup(mc.Finish)

	tests := []struct {
		name           string
		args           args
		want           int64
		err            error
		repositoryMock repositoryMockFunc
		lomsMock       lomsCallerMockFunc
	}{
		{
			name: "positive case",
			args: args{
				ctx:  ctx,
				user: user,
			},
			want: orderID,
			err:  nil,
			repositoryMock: func(mc *minimock.Controller) CartsRepository {
				mock := NewCartsRepositoryMock(t)
				mock.GetCartMock.Expect(ctx, user).Return(cartItems, nil)
				mock.DeleteCartMock.Expect(ctx, user).Return(nil)
				return mock
			},
			lomsMock: func(mc *minimock.Controller) LOMSCaller {
				mock := NewLOMSCallerMock(t)
				mock.CreateOrderMock.Expect(ctx, user, cartItems).Return(lomsRes, nil)
				return mock
			},
		},
		{
			name: "negative case - repository error on get",
			args: args{
				ctx:  ctx,
				user: user,
			},
			want: orderIDError,
			err:  repoErr,
			repositoryMock: func(mc *minimock.Controller) CartsRepository {
				mock := NewCartsRepositoryMock(t)
				mock.GetCartMock.Expect(ctx, user).Return(nil, repoErr)
				return mock
			},
			lomsMock: func(mc *minimock.Controller) LOMSCaller {
				mock := NewLOMSCallerMock(t)
				return mock
			},
		},
		{
			name: "negative case - empty cart",
			args: args{
				ctx:  ctx,
				user: user,
			},
			want: orderIDError,
			err:  ErrNotItemsInCart,
			repositoryMock: func(mc *minimock.Controller) CartsRepository {
				mock := NewCartsRepositoryMock(t)
				mock.GetCartMock.Expect(ctx, user).Return(emptyCart, nil)
				return mock
			},
			lomsMock: func(mc *minimock.Controller) LOMSCaller {
				mock := NewLOMSCallerMock(t)
				return mock
			},
		},
		{
			name: "negative case - loms error",
			args: args{
				ctx:  ctx,
				user: user,
			},
			want: orderIDError,
			err:  lomsErr,
			repositoryMock: func(mc *minimock.Controller) CartsRepository {
				mock := NewCartsRepositoryMock(t)
				mock.GetCartMock.Expect(ctx, user).Return(cartItems, nil)
				return mock
			},
			lomsMock: func(mc *minimock.Controller) LOMSCaller {
				mock := NewLOMSCallerMock(t)
				mock.CreateOrderMock.Expect(ctx, user, cartItems).Return(lomsErrorRes, lomsErr)
				return mock
			},
		},
		{
			name: "negative case - repository error on delete",
			args: args{
				ctx:  ctx,
				user: user,
			},
			want: orderID,
			err:  repoErr,
			repositoryMock: func(mc *minimock.Controller) CartsRepository {
				mock := NewCartsRepositoryMock(t)
				mock.GetCartMock.Expect(ctx, user).Return(cartItems, nil)
				mock.DeleteCartMock.Expect(ctx, user).Return(repoErr)
				return mock
			},
			lomsMock: func(mc *minimock.Controller) LOMSCaller {
				mock := NewLOMSCallerMock(t)
				mock.CreateOrderMock.Expect(ctx, user, cartItems).Return(lomsRes, nil)
				return mock
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			api, err := NewMock(
				tt.lomsMock(mc),
				tt.repositoryMock(mc),
			)
			if err != nil {
				require.Equal(t, nil, err)
			}
			res, err := api.Purchase(tt.args.ctx, tt.args.user)
			require.Equal(t, tt.want, res)
			if tt.err != nil {
				require.ErrorContains(t, err, tt.err.Error())
			} else {
				require.Equal(t, tt.err, err)
			}
		})
	}

}
