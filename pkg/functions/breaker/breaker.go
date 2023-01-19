package breaker

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/trafficmanager/armtrafficmanager"
	"github.com/torumakabe/simple-circuit-breaker-for-azure-tm/pkg/alert"
	"github.com/torumakabe/simple-circuit-breaker-for-azure-tm/pkg/logger"
)

const (
	respAccepted = "Accepted"
)

var defaultTMBreaker Breaker = new(tmBreaker)

type Breaker interface {
	Break(subID, rgName, tmProfName string)
}

type tmBreaker struct{}

func (t *tmBreaker) Break(subID, rgName, tmProfName string) {
	defer func() {
		_ = logger.Sync()
	}()

	// for goroutine exec check
	logger.Debug("entered Break method")
	defer logger.Debug("finished Break method")

	// for unwrapping Azure RM API error code
	var respErr *azcore.ResponseError

	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		logger.Fatalf("failed to create Azure credential: %v", err)
	}

	// Azure Functions Consumption Plan has a timeout of 5 minutes. So, set it to 4 minutes considering the margin.
	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Minute)
	defer cancel()

	prof, err := getTMProfile(ctx, cred, subID, rgName, tmProfName)
	if err != nil {
		if errors.As(err, &respErr) {
			err = errors.New(respErr.ErrorCode)
		}
		logger.Errorf("failed to get Traffic Manager profile: %v", err)
		return
	}

	if string(*prof.Properties.TrafficRoutingMethod) != "Priority" {
		logger.Errorf("breaker will not trip bacause the routing method is not 'Priority': %v", string(*prof.Properties.TrafficRoutingMethod))
		return
	}

	if len(prof.Properties.Endpoints) < 2 {
		logger.Errorf("breaker will not trip due to low total endpoint count: %v", len(prof.Properties.Endpoints))
		return
	}

	var eps []*armtrafficmanager.Endpoint
	oc := 0
	for _, ep := range prof.Properties.Endpoints {
		if *ep.Properties.EndpointMonitorStatus == "Online" {
			oc++
		}
		logger.Infow("found endpoint",
			"name", *ep.Name,
			"priority", *ep.Properties.Priority,
			"status", *ep.Properties.EndpointStatus,
			"monitorStatus", *ep.Properties.EndpointMonitorStatus,
		)
		eps = append(eps, ep)
	}
	if oc < 1 {
		logger.Error("breaker will not trip due to no online endpoint")
		return
	}

	sortEndpoints(eps)

	for _, ep := range eps {
		if *ep.Properties.EndpointMonitorStatus == "Online" {
			logger.Infof("found an online and currently highest priority endpoint: %v. stop disabling endpoints", *ep.Name)
			return
		}

		if *ep.Properties.EndpointStatus != "Disabled" {
			logger.Infof("found an endpoint to disable: %v", *ep.Name)
			err := disableTMEndpoint(ctx, cred, subID, rgName, tmProfName, ep)
			if err != nil {
				if errors.As(err, &respErr) {
					err = errors.New(respErr.ErrorCode)
				}
				logger.Errorf("failed to disable endpoint: %v", err)
				return
			}
			logger.Infof("disable endpoint successfully: %v", *ep.Name)
		}
	}
}

func getTMProfile(ctx context.Context, cred azcore.TokenCredential, subID, rgName, profName string) (*armtrafficmanager.Profile, error) {
	client, err := armtrafficmanager.NewProfilesClient(subID, cred, nil)
	if err != nil {
		return nil, err
	}

	resp, err := client.Get(ctx, rgName, profName, nil)
	if err != nil {
		return nil, err
	}
	return &resp.Profile, nil
}

func sortEndpoints(eps []*armtrafficmanager.Endpoint) []*armtrafficmanager.Endpoint {
	sort.Slice(eps, func(i, j int) bool {
		return *eps[i].Properties.Priority < *eps[j].Properties.Priority
	})

	return eps
}

func disableTMEndpoint(ctx context.Context, cred azcore.TokenCredential, subID, rgName, profName string, ep *armtrafficmanager.Endpoint) error {
	client, err := armtrafficmanager.NewEndpointsClient(subID, cred, nil)
	if err != nil {
		return err
	}

	_, err = client.Update(
		ctx,
		rgName,
		profName,
		armtrafficmanager.EndpointType(strings.TrimPrefix(*ep.Type, "Microsoft.Network/trafficManagerProfiles/")),
		*ep.Name,
		armtrafficmanager.Endpoint{
			Name: ep.Name,
			Type: ep.Type,
			ID:   ep.ID,
			Properties: &armtrafficmanager.EndpointProperties{
				EndpointStatus: to.Ptr(armtrafficmanager.EndpointStatusDisabled),
			},
		},
		nil)
	if err != nil {
		return err
	}
	return nil
}

func HandleBreaker(w http.ResponseWriter, r *http.Request) {
	defer func() {
		_ = logger.Sync()
	}()

	var reqBody alert.Payload
	err := json.NewDecoder(r.Body).Decode(&reqBody)
	if err != nil {
		m := fmt.Sprintf("failed to decode request: %v", err)
		logger.Error(m)
		http.Error(w, m, http.StatusBadRequest)
		return
	}

	if reqBody.Data.Essentials.MonitorCondition != "Fired" {
		m := fmt.Sprintf("breaker will not trip because the condition of the alert is not 'Fired': %v", reqBody.Data.Essentials.MonitorCondition)
		logger.Error(m)
		w.WriteHeader(http.StatusNoContent)
		fmt.Fprint(w, m)
		return
	}

	// format /subscriptions/<subscription ID>/resourcegroups/<resource group name>/providers/microsoft.network/trafficmanagerprofiles/<Traffic Manager Profile name>"
	elem := strings.Split(strings.TrimPrefix(reqBody.Data.Essentials.AlertTargetIDs[0], "/"), "/")
	if len(elem) != 8 || elem[6] != "trafficmanagerprofiles" {
		m := fmt.Sprintf("breaker will not trip bacause ID format is not for Traffic Manager profile: %v", reqBody.Data.Essentials.AlertTargetIDs[0])
		logger.Error(m)
		http.Error(w, m, http.StatusBadRequest)
		return
	}
	subID := elem[1]
	logger.Infof("target subscription ID: %v", subID)
	rgName := elem[3]
	logger.Infof("target resource group: %v", rgName)
	tmProfName := elem[7]
	logger.Infof("target Traffic Manager profile: %v", tmProfName)

	// async
	go defaultTMBreaker.Break(subID, rgName, tmProfName)

	w.WriteHeader(http.StatusAccepted)
	fmt.Fprint(w, respAccepted)
}
