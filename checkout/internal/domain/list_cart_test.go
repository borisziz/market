package domain

import (
	"context"
	"testing"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/gojuno/minimock/v3"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
)

func TestListCart(t *testing.T) {
	type repositoryMockFunc func(mc *minimock.Controller) CartsRepository
	type productsMockFunc func(mc *minimock.Controller) ProductServiceCaller
	type limiterMockFunc func(mc *minimock.Controller) Limiter

	type args struct {
		ctx  context.Context
		user int64
	}

	var (
		mc                = minimock.NewController(t)
		ctx               = context.Background()
		repoErr           = errors.New("repo error")
		productsErr       = errors.New("products error")
		user        int64 = 1
		skus              = make(SKUs)

		cartItemsWithoutInfo = []CartItem{
			{
				Sku:   4678816,
				Count: gofakeit.Uint16(),
			},
			{
				Sku:   1148162,
				Count: gofakeit.Uint16(),
			},
			{
				Sku:   5415913,
				Count: gofakeit.Uint16(),
			},
		}
		cartItems = make([]CartItem, 0, len(cartItemsWithoutInfo))
		//emptyCart = make([]CartItem, 0, 0)
	)
	for _, item := range cartItemsWithoutInfo {
		item.ProductInfo = ProductInfo{
			Name:  gofakeit.Dog(),
			Price: gofakeit.Uint32(),
		}
		cartItems = append(cartItems, item)
	}
	t.Cleanup(mc.Finish)

	tests := []struct {
		name           string
		args           args
		want           []CartItem
		err            error
		repositoryMock repositoryMockFunc
		productsMock   productsMockFunc
		limiterMock    limiterMockFunc
	}{
		{
			name: "positive case",
			args: args{
				ctx:  ctx,
				user: user,
			},
			want: cartItems,
			err:  nil,
			productsMock: func(mc *minimock.Controller) ProductServiceCaller {
				mock := NewProductServiceCallerMock(t)
				mock.GetSKUsMock.Expect(ctx).Return(skus, nil)
				mock.GetProductMock.Set(func(ctx context.Context, sku uint32) (ProductInfo, error) {
					for _, item := range cartItems {
						if item.Sku == sku {
							return item.ProductInfo, nil
						}
					}
					return ProductInfo{}, productsErr
				})
				return mock
			},
			repositoryMock: func(mc *minimock.Controller) CartsRepository {
				mock := NewCartsRepositoryMock(t)
				mock.GetCartMock.Expect(ctx, user).Return(cartItemsWithoutInfo, nil)
				return mock
			},
			limiterMock: func(mc *minimock.Controller) Limiter {
				mock := NewLimiterMock(t)
				mock.WaitMock.Expect(ctx).Return(nil)
				return mock
			},
		},
		{
			name: "negative case - products error",
			args: args{
				ctx:  ctx,
				user: user,
			},
			want: nil,
			err:  productsErr,
			productsMock: func(mc *minimock.Controller) ProductServiceCaller {
				mock := NewProductServiceCallerMock(t)
				mock.GetSKUsMock.Expect(ctx).Return(skus, nil)
				mock.GetProductMock.Set(func(ctx context.Context, sku uint32) (ProductInfo, error) {
					return ProductInfo{}, productsErr
				})
				return mock
			},
			repositoryMock: func(mc *minimock.Controller) CartsRepository {
				mock := NewCartsRepositoryMock(t)
				mock.GetCartMock.Expect(ctx, user).Return(cartItemsWithoutInfo, nil)
				return mock
			},
			limiterMock: func(mc *minimock.Controller) Limiter {
				mock := NewLimiterMock(t)
				mock.WaitMock.Expect(ctx).Return(nil)
				return mock
			},
		},
		{
			name: "negative case - repository error",
			args: args{
				ctx:  ctx,
				user: user,
			},
			want: nil,
			err:  repoErr,
			productsMock: func(mc *minimock.Controller) ProductServiceCaller {
				mock := NewProductServiceCallerMock(t)
				mock.GetSKUsMock.Expect(ctx).Return(skus, nil)
				return mock
			},
			repositoryMock: func(mc *minimock.Controller) CartsRepository {
				mock := NewCartsRepositoryMock(t)
				mock.GetCartMock.Expect(ctx, user).Return(nil, repoErr)
				return mock
			},
			limiterMock: func(mc *minimock.Controller) Limiter {
				mock := NewLimiterMock(t)
				return mock
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			poolConfig := PoolConfig{
				AmountWorkers:     5,
				MaxRetries:        2,
				WithCancelOnError: true,
			}
			api, err := NewMock(
				tt.productsMock(mc),
				tt.repositoryMock(mc),
				tt.limiterMock(mc),
				poolConfig,
			)
			//tt.poolMock(mc)
			if err != nil {
				require.Equal(t, nil, err)
			}
			res, err := api.ListCart(tt.args.ctx, tt.args.user)
			require.Equal(t, tt.want, res)
			if tt.err != nil {
				require.ErrorContains(t, err, tt.err.Error())
			} else {
				require.Equal(t, tt.err, err)
			}
		})
	}

}
