package store

import (
	"context"
	"fmt"
	"math"
	"time"

	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/bitmark-inc/autonomy-api/schema"
)

var (
	ErrPOINotFound      = fmt.Errorf("poi not found")
	ErrPOIListNotFound  = fmt.Errorf("poi list not found")
	ErrPOIListMismatch  = fmt.Errorf("poi list mismatch")
	ErrProfileNotUpdate = fmt.Errorf("poi not update")
)

type POI interface {
	AddPOI(accountNumber string, alias, address string, lon, lat float64) (*schema.POI, error)
	ListPOI(accountNumber string) ([]*schema.POIDetail, error)

	GetPOI(poiID primitive.ObjectID) (*schema.POI, error)
	GetPOIMetrics(poiID primitive.ObjectID) (*schema.Metric, error)
	UpdatePOIMetric(poiID primitive.ObjectID, metric schema.Metric) error

	UpdatePOIAlias(accountNumber, alias string, poiID primitive.ObjectID) error
	UpdatePOIOrder(accountNumber string, poiOrder []string) error
	DeletePOI(accountNumber string, poiID primitive.ObjectID) error
	RefreshPOIState(poiID primitive.ObjectID, score float64) (bool, error)
}

// AddPOI inserts a new POI record if it doesn't exist and append it to user's profile
func (m *mongoDB) AddPOI(accountNumber string, alias, address string, lon, lat float64) (*schema.POI, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	c := m.client.Database(m.database).Collection(schema.POICollection)

	var poi schema.POI
	query := bson.M{
		"location.coordinates.0": lon,
		"location.coordinates.1": lat,
	}
	if err := c.FindOne(ctx, query).Decode(&poi); err != nil {
		if err == mongo.ErrNoDocuments {
			poi = schema.POI{
				Location: &schema.GeoJSON{
					Type:        "Point",
					Coordinates: []float64{lon, lat},
				},
			}
			result, err := c.InsertOne(ctx, bson.M{"location": poi.Location})
			if err != nil {
				return nil, err
			}
			poi.ID = result.InsertedID.(primitive.ObjectID)
		} else {
			return nil, err
		}
	}

	poiDesc := &schema.POIDesc{
		ID:      poi.ID,
		Alias:   alias,
		Address: address,
	}
	if err := m.AppendPOIToAccountProfile(accountNumber, poiDesc); err != nil {
		return nil, err
	}

	return &poi, nil
}

// ListPOI finds the POI list of an account along with customied alias and address
func (m *mongoDB) ListPOI(accountNumber string) ([]*schema.POIDetail, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	c := m.client.Database(m.database).Collection(schema.ProfileCollection)

	// find user's POI list
	var result struct {
		Points []*schema.POIDetail `bson:"points_of_interest"`
	}
	query := bson.M{"account_number": accountNumber}
	if err := c.FindOne(ctx, query).Decode(&result); err != nil {
		return nil, err
	}
	if result.Points == nil { // user hasn't tracked any location yet
		return []*schema.POIDetail{}, nil
	}

	// find scores
	poiIDs := make([]primitive.ObjectID, 0)
	for _, p := range result.Points {
		poiIDs = append(poiIDs, p.ID)
	}

	// $in query doesn't guarantee order
	// use aggregation to sort the nested docs according to the query order
	pipeline := []bson.M{
		{"$match": bson.M{"_id": bson.M{"$in": poiIDs}}},
		{"$addFields": bson.M{"__order": bson.M{"$indexOfArray": bson.A{poiIDs, "$_id"}}}},
		{"$sort": bson.M{"__order": 1}},
	}
	c = m.client.Database(m.database).Collection(schema.POICollection)
	cursor, err := c.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	var pois []*schema.POI
	if err = cursor.All(ctx, &pois); err != nil {
		return nil, err
	}

	if len(pois) != len(result.Points) {
		return nil, fmt.Errorf("poi data wrongly retrieved or removed")
	}

	for i, p := range result.Points {
		p.Score = pois[i].Metric.Score
		p.Location.Longitude = pois[i].Location.Coordinates[0]
		p.Location.Latitude = pois[i].Location.Coordinates[1]
	}

	return result.Points, nil
}

// GetPOI finds POI by poi ID
func (m *mongoDB) GetPOI(poiID primitive.ObjectID) (*schema.POI, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	c := m.client.Database(m.database).Collection(schema.POICollection)

	// find user's POI
	var poi *schema.POI
	query := bson.M{"_id": poiID}
	if err := c.FindOne(ctx, query).Decode(poi); err != nil {
		return nil, err
	}

	return poi, nil
}

// GetPOIMetrics finds POI by poi ID
func (m *mongoDB) GetPOIMetrics(poiID primitive.ObjectID) (*schema.Metric, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	c := m.client.Database(m.database).Collection(schema.POICollection)

	// find user's POI
	var poi schema.POI
	query := bson.M{"_id": poiID}
	if err := c.FindOne(ctx, query).Decode(&poi); err != nil {
		return nil, err
	}

	return &poi.Metric, nil
}

// UpdatePOIAlias updates the alias of a POI for the specified account
func (m *mongoDB) UpdatePOIAlias(accountNumber, alias string, poiID primitive.ObjectID) error {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	c := m.client.Database(m.database).Collection(schema.ProfileCollection)
	query := bson.M{
		"account_number":        accountNumber,
		"points_of_interest.id": poiID,
	}
	update := bson.M{"$set": bson.M{"points_of_interest.$.alias": alias}}
	result, err := c.UpdateOne(ctx, query, update)
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return ErrPOINotFound
	}

	return nil
}

// UpdatePOIOrder updates the order of the POIs for the specified account
func (m *mongoDB) UpdatePOIOrder(accountNumber string, poiOrder []string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	c := m.client.Database(m.database).Collection(schema.ProfileCollection)

	// construct mongodb aggregation $switch branches
	poiCondition := bson.A{}
	for i, id := range poiOrder {
		poiID, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			return err
		}

		poiCondition = append(poiCondition,
			bson.M{"case": bson.M{"$eq": bson.A{"$points_of_interest.id", poiID}}, "then": i})
	}

	cur, err := c.Aggregate(ctx, mongo.Pipeline{
		bson.D{
			{"$match", bson.D{
				{"account_number", accountNumber},
			}},
		},

		bson.D{
			{"$unwind", "$points_of_interest"},
		},

		bson.D{
			{"$addFields", bson.D{
				{"points_of_interest.order", bson.D{
					{"$switch", bson.D{{
						"branches", poiCondition},
					}},
				}},
			}},
		},

		bson.D{
			{"$sort", bson.M{
				"points_of_interest.order": 1}},
		},

		bson.D{
			{"$group", bson.D{
				{"_id", "$_id"},
				{"points_of_interest", bson.D{{"$push", "$points_of_interest"}}},
			}},
		},
	})

	if err != nil {
		switch e := err.(type) {
		case mongo.CommandError:
			if e.Code == 40066 { // $switch has no default and an input matched no case
				return ErrPOIListMismatch
			}
		default:
			return err
		}

	}

	var profiles []bson.M

	if err := cur.All(ctx, &profiles); nil != err {
		return err
	}

	if len(profiles) < 1 {
		return ErrPOIListNotFound
	}

	poi, ok := profiles[0]["points_of_interest"]
	if !ok {
		return ErrPOIListNotFound
	}

	query := bson.M{
		"account_number": accountNumber,
	}
	update := bson.M{"$set": bson.M{"points_of_interest": poi}}
	result, err := c.UpdateOne(ctx, query, update)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return ErrProfileNotUpdate
	}

	return nil
}

func (m *mongoDB) DeletePOI(accountNumber string, poiID primitive.ObjectID) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	c := m.client.Database(m.database).Collection(schema.ProfileCollection)
	query := bson.M{
		"account_number":        accountNumber,
		"points_of_interest.id": poiID,
	}
	update := bson.M{"$pull": bson.M{"points_of_interest": bson.M{"id": poiID}}}
	if _, err := c.UpdateOne(ctx, query, update); err != nil {
		return err
	}

	return nil
}

// RefreshPOIState checks current states of a specific POI and return true
// if the score has changed.
func (m mongoDB) RefreshPOIState(poiID primitive.ObjectID, score float64) (bool, error) {
	poi, err := m.GetPOI(poiID)
	if err != nil {
		return false, err
	}

	metric := poi.Metric
	changed := math.Abs(metric.Score-score) > 33

	// User current metric as new metric
	metric.Score = score
	metric.LastUpdate = time.Now().UTC().Unix()

	if err := m.UpdatePOIMetric(poiID, metric); err != nil {
		return false, err
	}

	return changed, nil
}

func (m *mongoDB) UpdatePOIMetric(poiID primitive.ObjectID, metric schema.Metric) error {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	c := m.client.Database(m.database).Collection(schema.POICollection)
	query := bson.M{
		"_id": poiID,
	}
	update := bson.M{
		"$set": bson.M{
			"metric": metric,
		},
	}

	result, err := c.UpdateOne(ctx, query, update)
	pid := poiID.String()
	if err != nil {
		log.WithFields(log.Fields{
			"prefix": mongoLogPrefix,
			"poi ID": pid,
			"error":  err,
		}).Error("update poi metric")
		return err
	}

	if result.MatchedCount == 0 {
		log.WithFields(log.Fields{
			"prefix": mongoLogPrefix,
			"poi ID": pid,
			"error":  ErrPOINotFound.Error(),
		}).Error("update poi metric")
		return ErrPOINotFound
	}

	return nil
}

// GetPOIByID finds POI by poi ID
func (m *mongoDB) GetPOIByID(poiID primitive.ObjectID) (*schema.POI, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()
	c := m.client.Database(m.database).Collection(schema.POICollection)

	// find user's POI
	var poi schema.POI
	query := bson.M{"_id": poiID}
	if err := c.FindOne(ctx, query).Decode(&poi); err != nil {
		return nil, err
	}
	return &poi, nil
}

func (m *mongoDB) NearestPOI(distance int, cords schema.Location) ([]primitive.ObjectID, error) {
	query := distanceQuery(distance, cords)
	c := m.client.Database(m.database).Collection(schema.POICollection)
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	cur, err := c.Find(ctx, query)
	if nil != err {
		log.WithField("prefix", mongoLogPrefix).Errorf("query poi nearest distance with error: %s", err)
		return []primitive.ObjectID{}, fmt.Errorf("poi nearest distance query with error: %s", err)
	}
	var record schema.POI
	var POIs []primitive.ObjectID

	// iterate
	for cur.Next(ctx) {
		err = cur.Decode(&record)
		if nil != err {
			log.WithField("prefix", mongoLogPrefix).Infof("query nearest distance with error: %s", err)
			return []primitive.ObjectID{}, fmt.Errorf("nearest distance query decode record with error: %s", err)
		}
		POIs = append(POIs, record.ID)
	}

	log.WithField("prefix", mongoLogPrefix).Debugf("poi nearest distance query gets %d records near long:%v lat:%v", len(POIs),
		cords.Longitude, cords.Latitude)

	return POIs, nil
}
