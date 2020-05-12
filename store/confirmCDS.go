package store

import (
	"context"
	"errors"
	"fmt"

	"github.com/bitmark-inc/autonomy-api/schema"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type CDSCountryType string

const (
	CdsUSA    = "United State"
	CdsTaiwan = "Taiwan"
)

var CDSCountyCollectionMatrix = map[CDSCountryType]string{
	CDSCountryType(CdsUSA):    "ConfirmUS",
	CDSCountryType(CdsTaiwan): "ConfirmTaiwan",
}

func (m *mongoDB) CreateCDSData(result []schema.CDSData, country string) error {
	collection, ok := CDSCountyCollectionMatrix[CDSCountryType(country)]
	if !ok {
		return errors.New("no cds country availible")
	}
	data := make([]interface{}, len(result))
	for i, v := range result {
		data[i] = v
	}
	opts := options.InsertMany().SetOrdered(false)
	res, err := m.client.Database(m.database).Collection(collection).InsertMany(context.Background(), data, opts)
	if err != nil {
		if errs, hasErr := err.(mongo.BulkWriteException); hasErr {
			if 1 == len(errs.WriteErrors) && DuplicateKeyCode == errs.WriteErrors[0].Code {
				fmt.Println(err)
				return nil
			}
		}
	}
	if res != nil {
		log.WithFields(log.Fields{
			"prefix":  mongoLogPrefix,
			"records": len(res.InsertedIDs),
		}).Info("createCDSData Insert data")
	}
	return nil
}
