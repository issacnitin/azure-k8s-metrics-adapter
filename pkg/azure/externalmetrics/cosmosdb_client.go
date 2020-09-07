package externalmetrics

import (
	"context"
	"strconv"

	"github.com/Azure/azure-sdk-for-go/services/cosmos-db/mgmt/2015-04-08/documentdb"

	"github.com/Azure/go-autorest/autorest/azure/auth"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"k8s.io/klog"
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

	klog.V(0).Infof("%s", azMetricRequest)

	connectionStrings, err := c.client.ListConnectionStrings(context.Background(), azMetricRequest.ResourceGroup, azMetricRequest.ResourceName)
	if err != nil || len(*connectionStrings.ConnectionStrings) == 0 {
		return AzureExternalMetricResponse{}, err
	}

	clientOptions := options.Client().ApplyURI(*(*connectionStrings.ConnectionStrings)[0].ConnectionString).SetDirect(true)
	cl, err := mongo.NewClient(clientOptions)
	if err != nil {
		return AzureExternalMetricResponse{}, err
	}
	err = cl.Connect(context.Background())

	if err != nil {
		return AzureExternalMetricResponse{}, err
	}

	var dict Doc
	col := cl.Database(azMetricRequest.DatabaseName).Collection(azMetricRequest.CollectionName)
	objID, err := primitive.ObjectIDFromHex(azMetricRequest.DocumentId)
	if err != nil {
		return AzureExternalMetricResponse{}, err
	}
	err = col.FindOne(context.Background(), bson.M{"_id": objID}).Decode(&dict)
	if err != nil {
		return AzureExternalMetricResponse{}, err
	}

	v, err := strconv.Atoi(dict.Available)
	if err != nil {
		return AzureExternalMetricResponse{}, err
	}
	return AzureExternalMetricResponse{
		Value: float64(v),
	}, err
}

type Doc struct {
	Available string      `json:"available"`
	Id        interface{} `json:"id" bson:"_id,omitempty"`
}
