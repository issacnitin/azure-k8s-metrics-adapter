package externalmetrics

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	"k8s.io/klog"
)

type AzureExternalMetricRequest struct {
	MetricName                string
	SubscriptionID            string
	Type                      string
	ResourceName              string
	ResourceProviderNamespace string
	ResourceType              string
	Aggregation               string
	Timespan                  string
	Filter                    string
	ResourceGroup             string
	Namespace                 string
	Topic                     string
	Subscription              string
	DatabaseName              string
	CollectionName            string
	DocumentId                string
	DocumentField             string
}

func ParseAzureMetric(metricSelector labels.Selector, defaultSubscriptionID string) (AzureExternalMetricRequest, error) {
	klog.V(4).Infof("Parsing a received AzureMetric")
	klog.V(6).Infof("%v", metricSelector)

	if metricSelector == nil {
		return AzureExternalMetricRequest{}, fmt.Errorf("metricSelector cannot be nil")
	}

	// Using selectors to pass required values thorugh
	// to retain camel case as azure provider is case sensitive.
	//
	// There is are restrictions so using some conversion
	// restrictions here
	// note: requirement values are already validated by apiserver
	merticReq := AzureExternalMetricRequest{
		Timespan:       TimeSpan(),
		SubscriptionID: defaultSubscriptionID,
	}

	requirements, _ := metricSelector.Requirements()
	for _, request := range requirements {
		if request.Operator() != selection.Equals {
			return AzureExternalMetricRequest{}, errors.New("selector type not supported. only equals is supported at this time")
		}

		value := request.Values().List()[0]
		klog.V(0).Infof("Checking key %s with value %s", request.Key(), value)
		switch request.Key() {
		// Shared
		case "metricName":
			klog.V(4).Infof("AzureMetric metricName: %s", value)
			merticReq.MetricName = value
		case "resourceGroup":
			klog.V(4).Infof("AzureMetric resourceGroup: %s", value)
			merticReq.ResourceGroup = value
		case "subscriptionID":
			// if sub id is passed via label selectors then it takes precedence
			klog.V(4).Infof("AzureMetric override azure subscription id with : %s", value)
			merticReq.SubscriptionID = value
		// Monitor
		case "resourceName":
			klog.V(2).Infof("resourceName: %s", value)
			merticReq.ResourceName = value
		case "resourceProviderNamespace":
			klog.V(2).Infof("resourceProviderNamespace: %s", value)
			merticReq.ResourceProviderNamespace = value
		case "resourceType":
			klog.V(2).Infof("resourceType: %s", value)
			merticReq.ResourceType = value
		case "aggregation":
			klog.V(2).Infof("aggregation: %s", value)
			merticReq.Aggregation = value
		case "filter":
			// TODO: Should handle filters by converting equality and setbased label selectors
			// to  oData syntax: https://kubernetes.io/docs/concepts/overview/working-with-objects/labels/#label-selectors
			klog.V(2).Infof("filter: %s", value)
			filterStrings := strings.Split(value, "_")
			merticReq.Filter = fmt.Sprintf("%s %s '%s'", filterStrings[0], filterStrings[1], filterStrings[2])
			klog.V(2).Infof("filter formatted: %s", merticReq.Filter)
		// Service Bus
		case "namespace":
			klog.V(4).Infof("AzureMetric namespace: %s", value)
			merticReq.Namespace = value
		case "topic":
			klog.V(4).Infof("AzureMetric topic: %s", value)
			merticReq.Topic = value
		case "subscription":
			klog.V(4).Infof("AzureMetric subscription: %s", value)
			merticReq.Subscription = value
		// Cosmos DB
		case "database":
			klog.V(2).Infof("AzureMetric database: %s", value)
			merticReq.DatabaseName = value
		case "collection":
			klog.V(2).Infof("AzureMetric collection: %s", value)
			merticReq.CollectionName = value
		case "documentid":
			klog.V(2).Infof("AzureMetric documentid: %s", value)
			merticReq.DocumentId = value
		case "documentfield":
			klog.V(2).Infof("AzureMetric documentfield: %s", value)
			merticReq.DocumentField = value
		default:
			return AzureExternalMetricRequest{}, fmt.Errorf("selector label '%s' not supported", request.Key())
		}
	}

	klog.V(2).Infof("Successfully parsed AzureMetric %s", merticReq.MetricName)

	return merticReq, nil
}

type InvalidMetricRequestError struct {
	err string
}

func (i InvalidMetricRequestError) Error() string {
	return fmt.Sprintf(i.err)
}

func IsInvalidMetricRequestError(err error) bool {
	if _, ok := err.(InvalidMetricRequestError); ok {
		return true
	}
	return false
}

func (amr AzureExternalMetricRequest) Validate() error {
	// Shared
	if amr.MetricName == "" {
		return InvalidMetricRequestError{err: "metricName is required"}
	}
	if amr.ResourceGroup == "" {
		return InvalidMetricRequestError{err: "resourceGroup is required"}
	}
	if amr.SubscriptionID == "" {
		return InvalidMetricRequestError{err: "subscriptionID is required. set a default or pass via label selectors"}
	}

	// Service Bus

	// if amr.Namespace == "" {
	// 	return InvalidMetricRequestError{err: "namespace is required"}
	// }
	// if amr.Topic == "" {
	// 	return InvalidMetricRequestError{err: "topic is required"}
	// }
	// if amr.Subscription == "" {
	// 	return InvalidMetricRequestError{err: "subscription is required"}
	// }

	// if here then valid!
	return nil
}

// TimeSpan sets the default time to aggregate a metric
func TimeSpan() string {
	// defaults to last five minutes.
	// TODO support configuration via config
	endtime := time.Now().UTC().Format(time.RFC3339)
	starttime := time.Now().Add(-(5 * time.Minute)).UTC().Format(time.RFC3339)
	return fmt.Sprintf("%s/%s", starttime, endtime)
}

func (amr AzureExternalMetricRequest) MetricResourceURI() string {
	return fmt.Sprintf("/subscriptions/%s/resourceGroups/%s/providers/%s/%s/%s",
		amr.SubscriptionID,
		amr.ResourceGroup,
		amr.ResourceProviderNamespace,
		amr.ResourceType,
		amr.ResourceName)
}
