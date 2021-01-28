package rbac

import (
	"testing"
	"time"

	"github.com/grafana/grafana/pkg/api/routing"
	"github.com/grafana/grafana/pkg/infra/log"
	"github.com/grafana/grafana/pkg/registry"
	"github.com/grafana/grafana/pkg/services/sqlstore"
	"github.com/grafana/grafana/pkg/setting"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestEnv(t *testing.T) *RBACService {
	cfg := setting.NewCfg()

	ac := overrideRBACInRegistry(cfg)

	sqlStore := sqlstore.InitTestDB(t)
	ac.SQLStore = sqlStore

	err := ac.Init()
	require.NoError(t, err)
	return &ac
}

func overrideRBACInRegistry(cfg *setting.Cfg) RBACService {
	ac := RBACService{
		SQLStore:      nil,
		Cfg:           cfg,
		RouteRegister: routing.NewRouteRegister(),
		log:           log.New("rbac-test"),
	}

	overrideServiceFunc := func(descriptor registry.Descriptor) (*registry.Descriptor, bool) {
		if _, ok := descriptor.Instance.(*RBACService); ok {
			return &registry.Descriptor{
				Name:         "RBAC",
				Instance:     &ac,
				InitPriority: descriptor.InitPriority,
			}, true
		}
		return nil, false
	}

	registry.RegisterOverride(overrideServiceFunc)

	return ac
}

func mockTimeNow() {
	var timeSeed int64
	timeNow = func() time.Time {
		fakeNow := time.Unix(timeSeed, 0).UTC()
		timeSeed++
		return fakeNow
	}
}

func resetTimeNow() {
	timeNow = time.Now
}

func TestCreatingPolicy(t *testing.T) {
	mockTimeNow()
	defer resetTimeNow()

	testCases := []struct {
		desc        string
		inputName   string
		permissions []struct {
			resource string
			action   string
		}

		expectedError   error
		expectedUpdated time.Time
	}{
		{
			desc:            "should successfuly create simple policy",
			inputName:       "a name",
			permissions:     nil,
			expectedUpdated: time.Unix(1, 0).UTC(),
		},
		{
			desc:      "should successfuly create policy with permissions",
			inputName: "a name",
			permissions: []struct {
				resource string
				action   string
			}{
				{resource: "/api/admin/users", action: "post"},
				{resource: "/api/report", action: "get"},
			},
			expectedUpdated: time.Unix(3, 0).UTC(),
		},
	}
	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			ac := setupTestEnv(t)
			t.Cleanup(registry.ClearOverrides)

			cmd := CreatePolicyCommand{
				OrgId: 1,
				Name:  tc.inputName,
			}

			err := ac.CreatePolicy(&cmd)
			if tc.expectedError != nil {
				require.Error(t, err)
				return
			}

			policyId := cmd.Result.Id

			if tc.permissions != nil {
				for _, p := range tc.permissions {
					permCmd := CreatePermissionCommand{
						OrgId:    1,
						PolicyId: policyId,
						Resource: p.resource,
						Action:   p.action,
					}

					err := ac.CreatePermission(&permCmd)
					require.NoError(t, err)
				}
			}

			q := GetPolicyQuery{
				OrgId:    1,
				PolicyId: policyId,
			}

			err = ac.GetPolicy(&q)
			policy := q.Result

			require.NoError(t, err)
			assert.Equal(t, tc.expectedUpdated, policy.Updated)
			if tc.permissions == nil {
				assert.Empty(t, policy.Permissions)
			} else {
				assert.Equal(t, len(tc.permissions), len(policy.Permissions))
				for i, p := range policy.Permissions {
					assert.Equal(t, tc.permissions[i].resource, p.Resource)
					assert.Equal(t, tc.permissions[i].action, p.Action)
				}
			}
		})
	}
}
