package externalmetrics

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/services/cosmos-db/mgmt/2020-04-01/documentdb"
	"github.com/Azure/azure-sdk-for-go/services/cosmos-db/mongodb/mongodb"
	"github.com/Azure/go-autorest/autorest/azure/auth"
	"go.mongodb.org/mongo-driver/bson"
)

type cosmosmonitorClient interface {
	documentdbapi.DatabaseAccountsClient
}

type cosmosClient struct {
	client                cosmosmonitorClient
	DefaultSubscriptionID string
}

func NewCosmosClient(defaultsubscriptionID string) AzureExternalMetricClient {
	client := documentdb.NewDatabaseAccountsClient(defaultsubscriptionID)
	authorizer, err := auth.NewAuthorizerFromEnvironment()
	if err == nil {
		client.Authorizer = authorizer
	}

	return &monitorClient{
		client:                client,
		DefaultSubscriptionID: defaultsubscriptionID,
	}
}

func newCosmosClient(defaultsubscriptionID string, client cosmosmonitorClient) monitorClient {
	return monitorClient{
		client:                client,
		DefaultSubscriptionID: defaultsubscriptionID,
	}
}

// GetAzureMetric calls Azure Monitor endpoint and returns a metric
func (c *monitorClient) GetAzureMetric(azMetricRequest AzureExternalMetricRequest) (AzureExternalMetricResponse, error) {
	err := azMetricRequest.Validate()
	if err != nil {
		return AzureExternalMetricResponse{}, err
	}

	connectionStrings, err := c.client.ListConnectionStrings(context.Background(), azMetricRequest.ResourceGroup, azMetricRequest.ResourceName)
	if err != nil || len(connectionStrings) == 0 {
		return AzureExternalMetricResponse{}, err
	}
	mongoClient, err := mongodb.NewMongoDBClientWithConnectionString(connectionStrings[0])
	if err != nil {
		return AzureExternalMetricResponse{}, err
	}

	con := mongoClient.DB(azMetricRequest.Database).C(azMetricRequest.Client)
	var dict map[string]string
	err = c.Find(bson.M{"_id": azMetricRequest.DocumentId}).One(&dict)
	if err != nil {
		return AzureExternalMetricResponse{}, err
	}

	if val, ok := dict[azMetricRequest.DocumentField]; ok {
		return AzureExternalMetricResponse{
			Value: inter[],
		}, err
	}

	return AzureExternalMetricResponse{}, err
}
