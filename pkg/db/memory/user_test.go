package memory

import (
	"testing"

	"github.com/sh-miyoshi/jwt-server/pkg/db/model"
)

func TestFilterUserList(t *testing.T) {
	tt := []struct {
		UserName   string
		FilterName string
		ExpectNum  int
	}{
		{
			UserName:   "admin",
			FilterName: "admin",
			ExpectNum:  1,
		},
		{
			UserName:   "admin",
			FilterName: "",
			ExpectNum:  1,
		},
		{
			UserName:   "admin",
			FilterName: "fakeadmin",
			ExpectNum:  0,
		},
		{
			UserName:   "admin",
			FilterName: "adminfake",
			ExpectNum:  0,
		},
	}

	for _, tc := range tt {
		data := []*model.UserInfo{
			&model.UserInfo{
				Name: tc.UserName,
			},
		}

		filter := &model.UserFilter{
			Name: tc.FilterName,
		}

		res := filterUserList(data, filter)
		if len(res) != tc.ExpectNum {
			t.Errorf("Filter User List Failed: expect num: %d, but got %d, %v", tc.ExpectNum, len(res), res)
		}
	}
}