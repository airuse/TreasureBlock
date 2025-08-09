package services

import (
	"blockChainBrowser/server/internal/models"
	"blockChainBrowser/server/internal/repository"
	"context"
	"fmt"
)

type AddressService interface {
	CreateAddress(ctx context.Context, address *models.Address) error
	GetAddressByAddress(ctx context.Context, address string) (*models.Address, error)
}

type addressService struct {
	addressRepo repository.AddressRepository
}

func NewAddressService(addressRepo repository.AddressRepository) AddressService {
	return &addressService{
		addressRepo: addressRepo,
	}
}

func (s *addressService) CreateAddress(ctx context.Context, address *models.Address) error {
	if address == nil {
		return fmt.Errorf("address cannot be nil")
	}
	if address.Address == "" {
		return fmt.Errorf("address cannot be empty")
	}
	return s.addressRepo.Create(ctx, address)
}

func (s *addressService) GetAddressByAddress(ctx context.Context, address string) (*models.Address, error) {
	if address == "" {
		return nil, fmt.Errorf("address cannot be empty")
	}
	addr, err := s.addressRepo.GetByAddress(ctx, address)
	if err != nil {
		return nil, fmt.Errorf("failed to get address by address: %w", err)
	}
	return addr, nil
}
