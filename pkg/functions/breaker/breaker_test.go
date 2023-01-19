package breaker_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/trafficmanager/armtrafficmanager"
	"github.com/torumakabe/simple-circuit-breaker-for-azure-tm/pkg/alert"
	"github.com/torumakabe/simple-circuit-breaker-for-azure-tm/pkg/functions/breaker"
)

type dummyBreak struct{}

func (m *dummyBreak) Break(subID, rgName, tmProfName string) {
	// do nothing
}

func Test_sortEndpoints(t *testing.T) {
	sortedResult := []*armtrafficmanager.Endpoint{
		{
			Name: to.Ptr("target1"),
			Type: to.Ptr("Microsoft.Network/trafficManagerProfiles/externalEndpoints"),
			ID:   to.Ptr("target1"),
			Properties: &armtrafficmanager.EndpointProperties{
				EndpointStatus:        to.Ptr(armtrafficmanager.EndpointStatusEnabled),
				EndpointMonitorStatus: to.Ptr(armtrafficmanager.EndpointMonitorStatusOnline),
				Priority:              to.Ptr(int64(1)),
				Target:                to.Ptr("target1.example.com"),
			},
		},
		{
			Name: to.Ptr("target2"),
			Type: to.Ptr("Microsoft.Network/trafficManagerProfiles/externalEndpoints"),
			ID:   to.Ptr("target2"),
			Properties: &armtrafficmanager.EndpointProperties{
				EndpointStatus:        to.Ptr(armtrafficmanager.EndpointStatusEnabled),
				EndpointMonitorStatus: to.Ptr(armtrafficmanager.EndpointMonitorStatusOnline),
				Priority:              to.Ptr(int64(2)),
				Target:                to.Ptr("target2.example.com"),
			},
		},
		{
			Name: to.Ptr("target3"),
			Type: to.Ptr("Microsoft.Network/trafficManagerProfiles/externalEndpoints"),
			ID:   to.Ptr("target3"),
			Properties: &armtrafficmanager.EndpointProperties{
				EndpointStatus:        to.Ptr(armtrafficmanager.EndpointStatusEnabled),
				EndpointMonitorStatus: to.Ptr(armtrafficmanager.EndpointMonitorStatusOnline),
				Priority:              to.Ptr(int64(3)),
				Target:                to.Ptr("target3.example.com"),
			},
		},
	}

	type args struct {
		eps []*armtrafficmanager.Endpoint
	}
	tests := []struct {
		name string
		args args
		want []*armtrafficmanager.Endpoint
	}{
		{
			"ascending",
			args{
				[]*armtrafficmanager.Endpoint{
					{
						Name: to.Ptr("target1"),
						Type: to.Ptr("Microsoft.Network/trafficManagerProfiles/externalEndpoints"),
						ID:   to.Ptr("target1"),
						Properties: &armtrafficmanager.EndpointProperties{
							EndpointStatus:        to.Ptr(armtrafficmanager.EndpointStatusEnabled),
							EndpointMonitorStatus: to.Ptr(armtrafficmanager.EndpointMonitorStatusOnline),
							Priority:              to.Ptr(int64(1)),
							Target:                to.Ptr("target1.example.com"),
						},
					},
					{
						Name: to.Ptr("target2"),
						Type: to.Ptr("Microsoft.Network/trafficManagerProfiles/externalEndpoints"),
						ID:   to.Ptr("target2"),
						Properties: &armtrafficmanager.EndpointProperties{
							EndpointStatus:        to.Ptr(armtrafficmanager.EndpointStatusEnabled),
							EndpointMonitorStatus: to.Ptr(armtrafficmanager.EndpointMonitorStatusOnline),
							Priority:              to.Ptr(int64(2)),
							Target:                to.Ptr("target2.example.com"),
						},
					},
					{
						Name: to.Ptr("target3"),
						Type: to.Ptr("Microsoft.Network/trafficManagerProfiles/externalEndpoints"),
						ID:   to.Ptr("target3"),
						Properties: &armtrafficmanager.EndpointProperties{
							EndpointStatus:        to.Ptr(armtrafficmanager.EndpointStatusEnabled),
							EndpointMonitorStatus: to.Ptr(armtrafficmanager.EndpointMonitorStatusOnline),
							Priority:              to.Ptr(int64(3)),
							Target:                to.Ptr("target3.example.com"),
						},
					},
				},
			},
			sortedResult,
		},
		{
			"descending",
			args{
				[]*armtrafficmanager.Endpoint{
					{
						Name: to.Ptr("target3"),
						Type: to.Ptr("Microsoft.Network/trafficManagerProfiles/externalEndpoints"),
						ID:   to.Ptr("target3"),
						Properties: &armtrafficmanager.EndpointProperties{
							EndpointStatus:        to.Ptr(armtrafficmanager.EndpointStatusEnabled),
							EndpointMonitorStatus: to.Ptr(armtrafficmanager.EndpointMonitorStatusOnline),
							Priority:              to.Ptr(int64(3)),
							Target:                to.Ptr("target3.example.com"),
						},
					},
					{
						Name: to.Ptr("target2"),
						Type: to.Ptr("Microsoft.Network/trafficManagerProfiles/externalEndpoints"),
						ID:   to.Ptr("target2"),
						Properties: &armtrafficmanager.EndpointProperties{
							EndpointStatus:        to.Ptr(armtrafficmanager.EndpointStatusEnabled),
							EndpointMonitorStatus: to.Ptr(armtrafficmanager.EndpointMonitorStatusOnline),
							Priority:              to.Ptr(int64(2)),
							Target:                to.Ptr("target2.example.com"),
						},
					},
					{
						Name: to.Ptr("target1"),
						Type: to.Ptr("Microsoft.Network/trafficManagerProfiles/externalEndpoints"),
						ID:   to.Ptr("target1"),
						Properties: &armtrafficmanager.EndpointProperties{
							EndpointStatus:        to.Ptr(armtrafficmanager.EndpointStatusEnabled),
							EndpointMonitorStatus: to.Ptr(armtrafficmanager.EndpointMonitorStatusOnline),
							Priority:              to.Ptr(int64(1)),
							Target:                to.Ptr("target1.example.com"),
						},
					},
				},
			},
			sortedResult,
		},
		{
			"irregular",
			args{
				[]*armtrafficmanager.Endpoint{
					{
						Name: to.Ptr("target2"),
						Type: to.Ptr("Microsoft.Network/trafficManagerProfiles/externalEndpoints"),
						ID:   to.Ptr("target2"),
						Properties: &armtrafficmanager.EndpointProperties{
							EndpointStatus:        to.Ptr(armtrafficmanager.EndpointStatusEnabled),
							EndpointMonitorStatus: to.Ptr(armtrafficmanager.EndpointMonitorStatusOnline),
							Priority:              to.Ptr(int64(2)),
							Target:                to.Ptr("target2.example.com"),
						},
					},
					{
						Name: to.Ptr("target1"),
						Type: to.Ptr("Microsoft.Network/trafficManagerProfiles/externalEndpoints"),
						ID:   to.Ptr("target1"),
						Properties: &armtrafficmanager.EndpointProperties{
							EndpointStatus:        to.Ptr(armtrafficmanager.EndpointStatusEnabled),
							EndpointMonitorStatus: to.Ptr(armtrafficmanager.EndpointMonitorStatusOnline),
							Priority:              to.Ptr(int64(1)),
							Target:                to.Ptr("target1.example.com"),
						},
					},
					{
						Name: to.Ptr("target3"),
						Type: to.Ptr("Microsoft.Network/trafficManagerProfiles/externalEndpoints"),
						ID:   to.Ptr("target3"),
						Properties: &armtrafficmanager.EndpointProperties{
							EndpointStatus:        to.Ptr(armtrafficmanager.EndpointStatusEnabled),
							EndpointMonitorStatus: to.Ptr(armtrafficmanager.EndpointMonitorStatusOnline),
							Priority:              to.Ptr(int64(3)),
							Target:                to.Ptr("target3.example.com"),
						},
					},
				},
			},
			sortedResult,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := breaker.SortEndpoints(tt.args.eps); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("want %v, but %v", tt.want, got)
			}
		})
	}
}

func Test_HandleBreaker(t *testing.T) {
	defer breaker.SetTMBreaker(new(dummyBreak))()

	type args struct {
		payload alert.Payload
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{
			"success",
			args{
				alert.Payload{
					SchemaID: "test",
					Data: alert.Data{
						Essentials: alert.Essentials{
							MonitorCondition: "Fired",
							AlertTargetIDs:   []string{"/subscriptions/11111111-1111-1111-1111-11111111/resourcegroups/rg-test/providers/microsoft.network/trafficmanagerprofiles/tmprof-test"},
						},
					},
				},
			},
			http.StatusAccepted,
		},
		{
			"unmatch monitor condition",
			args{
				alert.Payload{
					SchemaID: "test",
					Data: alert.Data{
						Essentials: alert.Essentials{
							MonitorCondition: "Resolved",
							AlertTargetIDs:   []string{"/subscriptions/11111111-1111-1111-1111-11111111/resourcegroups/rg-test/providers/microsoft.network/trafficmanagerprofiles/tmprof-test"},
						},
					},
				},
			},
			http.StatusNoContent,
		},
		{
			"unmatch resource type",
			args{
				alert.Payload{
					SchemaID: "test",
					Data: alert.Data{
						Essentials: alert.Essentials{
							MonitorCondition: "Fired",
							AlertTargetIDs:   []string{"/subscriptions/11111111-1111-1111-1111-11111111/resourceGroups/rg-test/providers/Microsoft.Storage/storageAccounts/tmprof-test"},
						},
					},
				},
			},
			http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonData, err := json.Marshal(tt.args.payload)
			if err != nil {
				t.Fatal(err)
			}

			reqBody := bytes.NewBuffer(jsonData)
			req := httptest.NewRequest(http.MethodPost, "/", reqBody)
			got := httptest.NewRecorder()
			breaker.HandleBreaker(got, req)
			if got.Code != tt.want {
				t.Errorf("want %d, but %d", tt.want, got.Code)
			}
		})
	}
}
