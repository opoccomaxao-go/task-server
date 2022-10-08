package storage

import (
	"context"
	"sync"
	"time"

	"github.com/opoccomaxao-go/task-server/notificator"
	"github.com/opoccomaxao-go/task-server/task"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

const defaultCollection = "tasks"

type StorageMongoConfig struct {
	ConnectURL string // mongodb connect url
	DBName     string // used db, default: task
}

type storageMongo struct {
	mu     sync.Mutex
	client *mongo.Client
	db     *mongo.Database
	table  *mongo.Collection
	cfg    StorageMongoConfig
	noty   *notificator.Storage
}

func NewMongo(cfg StorageMongoConfig) (task.Storage, error) {
	res := storageMongo{
		cfg: cfg,
	}

	return &res, res.init()
}

func (s *storageMongo) init() error {
	if err := s.connect(); err != nil {
		return err
	}

	s.db = s.client.Database(s.cfg.DBName)
	s.table = s.db.Collection(defaultCollection)

	if err := s.updateCollectionIndices(); err != nil {
		return err
	}

	s.noty = notificator.NewStorage()

	return nil
}

func (s *storageMongo) connect() error {
	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(s.cfg.ConnectURL))
	if err != nil {
		return errors.WithStack(err)
	}

	s.client = client

	if err := client.Ping(context.Background(), readpref.Primary()); err != nil {
		return errors.WithStack(err)
	}

	return nil
}

func (s *storageMongo) updateCollectionIndices() error {
	const idxName = "expiration"

	_, _ = s.table.Indexes().DropOne(
		context.Background(),
		idxName,
	)

	name, err := s.table.Indexes().CreateOne(
		context.Background(),
		mongo.IndexModel{
			Keys: bson.M{
				"expiration": 1,
			},
			Options: options.Index().
				SetExpireAfterSeconds(1).
				SetName(idxName),
		},
	)
	if err != nil {
		return errors.WithStack(err)
	}

	if name != idxName {
		return task.ErrDBInvalidIndex
	}

	return nil
}

func (s *storageMongo) Create(t task.Task) error {
	_, err := s.table.InsertOne(context.Background(), taskToMongoTask(t))

	go s.noty.NotifyAll()

	return errors.WithStack(err)
}

func (s *storageMongo) Update(t task.Task) error {
	_, err := s.table.UpdateByID(
		context.Background(),
		t.ID,
		bson.M{"$set": taskToMongoTask(t)},
	)

	go s.noty.NotifyAll()

	return errors.WithStack(err)
}

func (s *storageMongo) FirstToExecute() (*task.Task, error) {
	resp := s.table.FindOne(
		context.Background(),
		bson.M{"executed": false},
	)

	if err := resp.Err(); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, errors.WithStack(task.ErrEmpty)
		}

		return nil, errors.WithStack(err)
	}

	var res mongoTask

	if err := resp.Decode(&res); err != nil {
		return nil, errors.WithStack(task.ErrDBFailed)
	}

	return res.Task(), nil
}

func (s *storageMongo) Watch() (task.Notificator, error) {
	stream, err := s.table.Watch(context.Background(), mongo.Pipeline{})
	if err != nil {
		noty := notificator.New().NotifyEveryTick(time.Second * 30)
		s.noty.Add(noty)

		//nolint:errcheck // first call on new instance does not actually return error
		go noty.Notify()

		//lint:ignore nilerr // fallback case
		return noty, nil
	}

	return &iteratorMongo{
		stream: stream,
	}, nil
}

func (s *storageMongo) Close() error {
	_ = s.noty.Close()

	return errors.WithStack(s.client.Disconnect(context.Background()))
}
