package main

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// DB -
type DB struct {
	client   *mongo.Client
	products *mongo.Collection
}

// NewDB -
func NewDB(url string) (*DB, error) {
	clientOpts := options.Client().ApplyURI(url)
	clientOpts.SetMaxConnIdleTime(300 * time.Second)
	clientOpts.SetMaxPoolSize(100)
	clientOpts.SetMinPoolSize(10)

	client, err := mongo.Connect(context.TODO(), clientOpts)
	if err != nil {
		return nil, err
	}

	err = client.Ping(context.TODO(), nil)
	if err != nil {
		return nil, err
	}

	db := &DB{client: client, products: client.Database("atlant").Collection("products")}

	err = db.indexMaker()
	if err != nil {
		return nil, err
	}

	return db, nil
}

func (db *DB) indexMaker() error {
	_, err := db.products.Indexes().CreateMany(context.TODO(), []mongo.IndexModel{
		{Keys: bson.M{"name": 1}, Options: options.Index().SetUnique(true)},
		{Keys: bson.M{"price": 1}},
		{Keys: bson.M{"last_updated": 1}},
		{Keys: bson.M{"changes_count": 1}},
	}, options.CreateIndexes())
	return err
}

// Disconnect - close DB connection
func (db *DB) Disconnect() {
	db.client.Disconnect(context.TODO())
}

// Save - save product in to db
func (db *DB) Save(name string, price int64) error {
	_, err := db.products.UpdateOne(context.TODO(), bson.M{"name": name}, bson.M{
		"$setOnInsert": bson.M{"name": name},
		"$set": bson.M{
			"price":        price,
			"last_updated": time.Now().UTC(),
		},
		"$inc": bson.M{"changes_count": 1},
	}, options.Update().SetUpsert(true))
	if err != nil {
		return err
	}

	return nil
}

// Product - one product
type Product struct {
	Name         string     `bson:"name, omitempty"`
	Price        int64      `bson:"price, omitempty"`
	LastUpdated  *time.Time `bson:"last_updated, omitempty"`
	ChangesCount uint64     `bson:"changes_count, omitempty"`
}

// List - get list of products with sorting
func (db *DB) List(ctx context.Context, skip, resultPerPage uint32, sort ...bson.E) ([]*Product, error) {
	if resultPerPage <= 0 {
		return []*Product{}, errors.New("Result Per Page can't be less than 1")
	}

	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	findOpts := options.Find().SetLimit(int64(resultPerPage))
	findOpts.SetSkip(int64(skip))
	findOpts.SetSort(sort)

	cur, err := db.products.Find(ctx, bson.D{}, findOpts)
	if err != nil {
		return []*Product{}, err
	}
	defer cur.Close(ctx)

	var products = make([]*Product, 0, resultPerPage)

	for cur.Next(ctx) {
		var p Product
		err := cur.Decode(&p)
		if err != nil {
			return []*Product{}, err
		}

		products = append(products, &p)
	}

	err = cur.Err()
	if err != nil {
		return []*Product{}, err
	}

	return products, err
}

// PagesCount - count pages by items per page
func (db *DB) PagesCount(ctx context.Context, resultPerPage uint32) (uint32, error) {
	if resultPerPage <= 0 {
		return 0, errors.New("Items Per Page can't be less than 1")
	}

	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	count, err := db.products.CountDocuments(ctx, bson.D{})
	if err != nil {
		return 0, err
	}

	return uint32(count / int64(resultPerPage)), err
}
