package census

const VERSION = "0.0.1"

type QueryBuilder interface{}

type censusQueryBuilder struct {
	censusClient *CensusClient
}

func NewCensusQueryBuilder(censusClient *CensusClient) QueryBuilder {
	return &censusQueryBuilder{
		censusClient: censusClient,
	}
}
