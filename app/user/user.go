package user

import (
	"context"
	"errors"
	"github.com/rpccloud/rpc"
	"github.com/rpccloud/rpc/app/util"
	"github.com/rpccloud/rpc/internal"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"time"
)

var Service = rpc.NewServiceWithOnMount(
	func(service *internal.Service, data interface{}) error {
		if cfg, ok := data.(*util.MongoDatabaseConfig); !ok || cfg == nil {
			return errors.New("config error")
		} else if err := createUserIfNotExist(cfg); err != nil {
			return err
		} else {
			service.AddChildService("phone", phoneService, data)
			return nil
		}
	},
)

func createUserIfNotExist(cfg *util.MongoDatabaseConfig) error {
	return util.WithMongoClient(cfg.URI, 3*time.Second,
		func(client *mongo.Client, ctx context.Context) error {
			db := client.Database(cfg.DataBase)
			filter := bson.M{"name": "user"}
			if coll, err := db.ListCollectionNames(ctx, filter); err != nil {
				return err
			} else if len(coll) == 0 {
				return client.Database(cfg.DataBase).CreateCollection(ctx, "user")
			} else {
				return nil
			}
		},
	)
}
