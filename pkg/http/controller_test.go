package http

import (
	"database/sql"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/sqlitedialect"
	"github.com/uptrace/bun/driver/sqliteshim"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"testing"
)

//goland:noinspection ALL
func Test_applyFilter(t *testing.T) {
	tests := []struct {
		name        string
		filterName  string
		filterValue string
		model       interface{}
		expect      string
	}{
		{
			name:        "no filterName",
			filterName:  "",
			filterValue: "",
			model:       &MockModel{},
			expect:      "SELECT \"mock_models\".\"id\" FROM \"mock_models\" LIMIT 10",
		},
		{
			name:        "valid filterName",
			filterName:  "ids",
			filterValue: "1,2,3",
			model:       &MockModel{},
			expect:      "SELECT \"mock_models\".\"id\" FROM \"mock_models\" WHERE (id IN (1, 2, 3)) LIMIT 10",
		},
		{
			name:        "invalid filterName method",
			filterName:  "invalid",
			filterValue: "1,2,3",
			model:       &MockModel{},
			expect:      "SELECT \"mock_models\".\"id\" FROM \"mock_models\" LIMIT 10",
		},
	}

	sqldb, err := sql.Open(sqliteshim.ShimName, "file::memory:?cache=shared")
	if err != nil {
		panic(err)
	}

	db := bun.NewDB(sqldb, sqlitedialect.New())

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &gin.Context{}
			if tt.filterName != "" {
				c.Request = &http.Request{
					URL: &url.URL{
						RawQuery: "filter[" + tt.filterName + "]=" + url.QueryEscape(fmt.Sprintf("%v", tt.filterValue)),
					},
				}
			}

			m := MockModel{}

			query := db.NewSelect().
				Model(&m).
				Limit(10).
				Offset(0)

			ApplyFilter[MockModel](c, query)

			t.Log(query.String())

			assert.Equal(t, tt.expect, query.String())
		})
	}
}

type MockModel struct {
	bun.BaseModel `bun:"table:mock_models,alias:mock_models"`
	Id            int64 `bun:"id,pk,autoincrement" json:"id"`
}

func (t *MockModel) ByIds(value string, query *bun.SelectQuery) {
	ids := strings.Split(value, ",")

	int64Slice := make([]int64, 0, len(ids))
	for _, s := range ids {
		num, err := strconv.ParseInt(s, 10, 64) // base 10, 64-bit integer
		if err != nil {
			continue
		}
		int64Slice = append(int64Slice, num)
	}

	query = query.Where("id IN (?)", bun.In(int64Slice))
}
