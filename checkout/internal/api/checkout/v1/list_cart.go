package checkout

import (
	"context"
	desc "route256/checkout/pkg/checkout/v1"
)

func (i *Implementation) ListCart(ctx context.Context, req *desc.ListCartRequest) (*desc.ListCartResponse, error) {
	cartItems, err := i.checkoutService.ListCart(ctx, req.GetUser())
	if err != nil {
		return nil, err
	}
	items := make([]*desc.CartItem, 0, len(cartItems))
	var totalPrice uint32 = 0
	for _, item := range cartItems {
		items = append(items, &desc.CartItem{
			Sku:   item.Sku,
			Count: uint32(item.Count),
			Name:  item.Name,
			Price: item.Price,
		})
		totalPrice += item.Price * uint32(item.Count)
	}

	return &desc.ListCartResponse{
		Items:      items,
		TotalPrice: totalPrice,
	}, nil
}
