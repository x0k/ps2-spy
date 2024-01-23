package sanctuary

import (
	"github.com/x0k/ps2-spy/internal/lib/census2"
)

const CensusJSONField = "c:censusJSON"

const (
	Ns_ps2 = "ps2"
	Ns_pts = "pts"
)

const (
	GetQuery      = "get"
	CountQuery    = "count"
	DescribeQuery = "describe"
)

func NewQuery(queryType, namespace, collection string) *census2.Query {
	return census2.NewQuery(queryType, namespace, collection).
		Where(census2.Cond(CensusJSONField).Equals(census2.BoolWithDefaultTrue(false)))
}
