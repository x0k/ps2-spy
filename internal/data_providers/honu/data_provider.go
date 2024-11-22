package honu_data_provider

import "github.com/x0k/ps2-spy/internal/lib/honu"

type DataProvider struct {
	client *honu.Client
}

func New(client *honu.Client) *DataProvider {
	return &DataProvider{
		client: client,
	}
}
