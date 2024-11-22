package voidwell_data_provider

import "github.com/x0k/ps2-spy/internal/lib/voidwell"

type DataProvider struct {
	client *voidwell.Client
}

func New(client *voidwell.Client) *DataProvider {
	return &DataProvider{
		client: client,
	}
}
