package externalmetrics

import (
	"context"
	"strconv"

	"github.com/Azure/azure-sdk-for-go/services/cosmos-db/mgmt/2015-04-08/documentdb"
	"github.com/Azure/azure-sdk-for-go/services/cosmos-db/mongodb"
	"github.com/Azure/go-autorest/autorest/azure/auth"
	"go.mongodb.org/mongo-driver/bson"
)

type cosmosClient struct {
	client                documentdb.DatabaseAccountsClient
	DefaultSubscriptionID string
}

func NewCosmosClient(defaultsubscriptionID string) AzureExternalMetricClient {
	client := documentdb.NewDatabaseAccountsClient(defaultsubscriptionID)
	authorizer, err := auth.NewAuthorizerFromEnvironment()
	if err == nil {
		client.Authorizer = authorizer
	}

	return &cosmosClient{
		client:                client,
		DefaultSubscriptionID: defaultsubscriptionID,
	}
}

func newCosmosClient(defaultsubscriptionID string, client documentdb.DatabaseAccountsClient) cosmosClient {
	return cosmosClient{
		client:                client,
		DefaultSubscriptionID: defaultsubscriptionID,
	}
}

// GetAzureMetric calls Azure Monitor endpoint and returns a metric
func (c *cosmosClient) GetAzureMetric(azMetricRequest AzureExternalMetricRequest) (AzureExternalMetricResponse, error) {
	err := azMetricRequest.Validate()
	if err != nil {
		return AzureExternalMetricResponse{}, err
	}

	connectionStrings, err := c.client.ListConnectionStrings(context.Background(), azMetricRequest.ResourceGroup, azMetricRequest.ResourceName)
	if err != nil || len(*connectionStrings.ConnectionStrings) == 0 {
		return AzureExternalMetricResponse{}, err
	}
	mongoClient, err := mongodb.NewMongoDBClientWithConnectionString((*(*connectionStrings.ConnectionStrings)[0].ConnectionString))
	if err != nil {
		return AzureExternalMetricResponse{}, err
	}

	con := mongoClient.DB(azMetricRequest.DatabaseName).C(azMetricRequest.CollectionName)
	var dict map[string]string
	err = con.Find(bson.M{"_id": azMetricRequest.DocumentId}).One(&dict)
	if err != nil {
		return AzureExternalMetricResponse{}, err
	}

	if val, ok := dict[azMetricRequest.DocumentField]; ok {
		v, err := strconv.Atoi(val)
		if err != nil {
			return AzureExternalMetricResponse{}, err
		}
		return AzureExternalMetricResponse{
			Value: float64(v),
		}, err
	}

	return AzureExternalMetricResponse{}, err
}
