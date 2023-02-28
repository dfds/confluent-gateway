package models

import (
	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestClusterAccess_GetAclPendingCreation(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name string
		acl  []AclEntry
		want int
	}{
		{
			name: "null list",
			acl:  nil,
			want: 0,
		},
		{
			name: "empty list",
			acl:  []AclEntry{},
			want: 0,
		},
		{
			name: "single created",
			acl:  []AclEntry{{CreatedAt: &now}},
			want: 0,
		},
		{
			name: "multiple created",
			acl:  []AclEntry{{CreatedAt: &now}, {CreatedAt: &now}},
			want: 0,
		},
		{
			name: "single pending",
			acl:  []AclEntry{{}},
			want: 1,
		},
		{
			name: "multiple pending",
			acl:  []AclEntry{{}, {}},
			want: 2,
		},
		{
			name: "mixed",
			acl:  []AclEntry{{}, {CreatedAt: &now}, {}},
			want: 2,
		},
		{
			name: "all",
			acl:  createAclEntries("some-cap", uuid.NewV4()),
			want: 18,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ca := &ClusterAccess{
				Acl: tt.acl,
			}
			actual := len(ca.GetAclPendingCreation())
			assert.Equalf(t, tt.want, actual, "expected %d acl entries pending creation, got %d", tt.want, actual)
		})
	}
}
