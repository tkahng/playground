package core

import (
	"context"
	"log"

	"github.com/stephenafamo/bob"
	"github.com/tkahng/authgo/internal/repository"
	"github.com/tkahng/authgo/internal/tools/payment"
)

type StripeService struct {
	client *payment.StripeClient
}

func NewStripeService(client *payment.StripeClient) *StripeService {
	return &StripeService{client: client}
}

func (srv *StripeService) FindAndUpsertAllCustomers(ctx context.Context, exec bob.Executor) error {
	customers, err := srv.client.FindAllCustomers()
	if err != nil {
		log.Printf("error finding all customers: %s", err)
		return err
	}

	for _, customer := range customers {
		user, err := repository.GetUserByEmail(ctx, exec, customer.Email)
		if err != nil {
			log.Printf("error finding user by email: %s", customer.Email)
			continue
		}
		err = repository.UpsertCustomer(ctx, exec, user.ID, customer.ID)
		if err != nil {
			log.Printf("error upserting customer: %s", customer.Email)
			continue
		}
	}
	return nil
}

func (srv *StripeService) FindAndUpsertAllProducts(ctx context.Context, exec bob.Executor) error {
	products, err := srv.client.FindAllProducts()
	if err != nil {
		log.Printf("error finding all products: %s", err)
		return err
	}
	for _, product := range products {
		err = repository.UpsertProductFromStripe(ctx, exec, product)
		if err != nil {
			log.Printf("error upserting product: %s", product.ID)
			continue
		}
	}
	return nil
}

func (srv *StripeService) FindAndUpsertAllPrices(ctx context.Context, exec bob.Executor) error {
	prices, err := srv.client.FindAllPrices()
	if err != nil {
		log.Printf("error finding all prices: %s", err)
		return err
	}
	for _, price := range prices {
		err = repository.UpsertPriceFromStripe(ctx, exec, price)
		if err != nil {
			log.Printf("error upserting price: %s", price.ID)
			continue
		}
	}
	return nil
}

func (srv *StripeService) FindAndUpsertAllSubscriptions(ctx context.Context, exec bob.Executor) error {
	subs, err := srv.client.FindAllSubscriptions()
	if err != nil {
		log.Printf("error finding all subscriptions: %s", err)
		return err
	}
	for _, sub := range subs {
		customer, err := repository.FindCustomerByStripeId(ctx, exec, sub.Customer.ID)
		if err != nil {
			log.Printf("error finding customer: %s", sub.Customer.ID)
			continue
		}
		if customer == nil {
			log.Printf("customer not found: %s", sub.Customer.ID)
			continue
		}
		err = repository.UpsertSubscriptionFromStripe(ctx, exec, sub, customer.ID)
		if err != nil {
			log.Printf("error upserting subscription %s", sub.ID)
			continue
		}
	}
	return nil
}
