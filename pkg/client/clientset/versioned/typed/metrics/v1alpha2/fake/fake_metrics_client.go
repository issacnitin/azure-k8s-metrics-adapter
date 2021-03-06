// Code generated by client-gen. DO NOT EDIT.

package fake

import (
	v1alpha2 "github.com/Azure/azure-k8s-metrics-adapter/pkg/client/clientset/versioned/typed/metrics/v1alpha2"
	rest "k8s.io/client-go/rest"
	testing "k8s.io/client-go/testing"
)

type FakeAzureV1alpha2 struct {
	*testing.Fake
}

func (c *FakeAzureV1alpha2) CustomMetrics(namespace string) v1alpha2.CustomMetricInterface {
	return &FakeCustomMetrics{c, namespace}
}

func (c *FakeAzureV1alpha2) ExternalMetrics(namespace string) v1alpha2.ExternalMetricInterface {
	return &FakeExternalMetrics{c, namespace}
}

// RESTClient returns a RESTClient that is used to communicate
// with API server by this client implementation.
func (c *FakeAzureV1alpha2) RESTClient() rest.Interface {
	var ret *rest.RESTClient
	return ret
}
