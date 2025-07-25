package database

import (
	"context"
	"time"

	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/mongo-go-driver/bson"
	"github.com/pritunl/mongo-go-driver/mongo"
	"github.com/pritunl/mongo-go-driver/mongo/options"
	"github.com/pritunl/mongo-go-driver/mongo/readconcern"
	"github.com/pritunl/mongo-go-driver/mongo/writeconcern"
	"github.com/pritunl/mongo-go-driver/x/mongo/driver/connstring"
	"github.com/pritunl/pritunl-cloud/config"
	"github.com/pritunl/pritunl-cloud/constants"
	"github.com/pritunl/pritunl-cloud/errortypes"
	"github.com/pritunl/pritunl-cloud/requires"
	"github.com/sirupsen/logrus"
)

type Database struct {
	ctx      context.Context
	client   *mongo.Client
	database *mongo.Database
}

func (d *Database) Deadline() (time.Time, bool) {
	if d.ctx != nil {
		return d.ctx.Deadline()
	}
	return time.Time{}, false
}

func (d *Database) Done() <-chan struct{} {
	if d.ctx != nil {
		return d.ctx.Done()
	}
	return nil
}

func (d *Database) Err() error {
	if d.ctx != nil {
		return d.ctx.Err()
	}
	return nil
}

func (d *Database) Value(key interface{}) interface{} {
	if d.ctx != nil {
		return d.ctx.Value(key)
	}
	return nil
}

func (d *Database) String() string {
	return "context.database"
}

func (d *Database) Close() {
}

func (d *Database) GetCollection(name string) (coll *Collection) {
	coll = &Collection{
		db:         d,
		Collection: d.database.Collection(name),
	}
	return
}

func (d *Database) getCollectionWeak(name string) (coll *Collection) {
	opts := &options.CollectionOptions{}

	opts.WriteConcern = writeconcern.New(
		writeconcern.W(1),
		writeconcern.WTimeout(10*time.Second),
	)
	opts.ReadConcern = readconcern.Local()

	coll = &Collection{
		db:         d,
		Collection: d.database.Collection(name, opts),
	}
	return
}

func (d *Database) Users() (coll *Collection) {
	coll = d.GetCollection("users")
	return
}

func (d *Database) Policies() (coll *Collection) {
	coll = d.GetCollection("policies")
	return
}

func (d *Database) Devices() (coll *Collection) {
	coll = d.GetCollection("devices")
	return
}

func (d *Database) Alerts() (coll *Collection) {
	coll = d.GetCollection("alerts")
	return
}

func (d *Database) AlertsEvent() (coll *Collection) {
	coll = d.GetCollection("alerts_event")
	return
}

func (d *Database) AlertsEventLock() (coll *Collection) {
	coll = d.GetCollection("alerts_event_lock")
	return
}

func (d *Database) Pods() (coll *Collection) {
	coll = d.GetCollection("pods")
	return
}

func (d *Database) Units() (coll *Collection) {
	coll = d.GetCollection("units")
	return
}

func (d *Database) Specs() (coll *Collection) {
	coll = d.GetCollection("specs")
	return
}

func (d *Database) Deployments() (coll *Collection) {
	coll = d.GetCollection("deployments")
	return
}

func (d *Database) Sessions() (coll *Collection) {
	coll = d.GetCollection("sessions")
	return
}

func (d *Database) Tasks() (coll *Collection) {
	coll = d.GetCollection("tasks")
	return
}

func (d *Database) Tokens() (coll *Collection) {
	coll = d.GetCollection("tokens")
	return
}

func (d *Database) CsrfTokens() (coll *Collection) {
	coll = d.GetCollection("csrf_tokens")
	return
}

func (d *Database) SecondaryTokens() (coll *Collection) {
	coll = d.GetCollection("secondary_tokens")
	return
}

func (d *Database) Nonces() (coll *Collection) {
	coll = d.GetCollection("nonces")
	return
}

func (d *Database) Rokeys() (coll *Collection) {
	coll = d.GetCollection("rokeys")
	return
}

func (d *Database) Schedulers() (coll *Collection) {
	coll = d.GetCollection("schedulers")
	return
}

func (d *Database) Settings() (coll *Collection) {
	coll = d.GetCollection("settings")
	return
}

func (d *Database) Events() (coll *Collection) {
	coll = d.getCollectionWeak("events")
	return
}

func (d *Database) Nodes() (coll *Collection) {
	coll = d.GetCollection("nodes")
	return
}

func (d *Database) NodePorts() (coll *Collection) {
	coll = d.GetCollection("node_ports")
	return
}

func (d *Database) Organizations() (coll *Collection) {
	coll = d.GetCollection("organizations")
	return
}

func (d *Database) Storages() (coll *Collection) {
	coll = d.GetCollection("storages")
	return
}

func (d *Database) Images() (coll *Collection) {
	coll = d.GetCollection("images")
	return
}

func (d *Database) Datacenters() (coll *Collection) {
	coll = d.GetCollection("datacenters")
	return
}

func (d *Database) Zones() (coll *Collection) {
	coll = d.GetCollection("zones")
	return
}

func (d *Database) Shapes() (coll *Collection) {
	coll = d.GetCollection("shapes")
	return
}

func (d *Database) Balancers() (coll *Collection) {
	coll = d.GetCollection("balancers")
	return
}

func (d *Database) Instances() (coll *Collection) {
	coll = d.GetCollection("instances")
	return
}

func (d *Database) Pools() (coll *Collection) {
	coll = d.GetCollection("pools")
	return
}

func (d *Database) Disks() (coll *Collection) {
	coll = d.GetCollection("disks")
	return
}

func (d *Database) Blocks() (coll *Collection) {
	coll = d.GetCollection("blocks")
	return
}

func (d *Database) BlocksIp() (coll *Collection) {
	coll = d.GetCollection("blocks_ip")
	return
}

func (d *Database) LvmLock() (coll *Collection) {
	coll = d.GetCollection("lvm_lock")
	return
}

func (d *Database) Journal() (coll *Collection) {
	coll = d.GetCollection("journal")
	return
}

func (d *Database) Firewalls() (coll *Collection) {
	coll = d.GetCollection("firewalls")
	return
}

func (d *Database) Versions() (coll *Collection) {
	coll = d.GetCollection("versions")
	return
}

func (d *Database) Plans() (coll *Collection) {
	coll = d.GetCollection("plans")
	return
}

func (d *Database) Vpcs() (coll *Collection) {
	coll = d.GetCollection("vpcs")
	return
}

func (d *Database) VpcsIp() (coll *Collection) {
	coll = d.GetCollection("vpcs_ip")
	return
}

func (d *Database) Authorities() (coll *Collection) {
	coll = d.GetCollection("authorities")
	return
}

func (d *Database) Certificates() (coll *Collection) {
	coll = d.GetCollection("certificates")
	return
}

func (d *Database) Secrets() (coll *Collection) {
	coll = d.GetCollection("secrets")
	return
}

func (d *Database) Domains() (coll *Collection) {
	coll = d.GetCollection("domains")
	return
}

func (d *Database) DomainsRecords() (coll *Collection) {
	coll = d.GetCollection("domains_records")
	return
}

func (d *Database) AcmeChallenges() (coll *Collection) {
	coll = d.GetCollection("acme_challenges")
	return
}

func (d *Database) Logs() (coll *Collection) {
	coll = d.GetCollection("logs")
	return
}

func (d *Database) Audits() (coll *Collection) {
	coll = d.GetCollection("audits")
	return
}

func (d *Database) Geo() (coll *Collection) {
	coll = d.GetCollection("geo")
	return
}

func Connect() (err error) {
	mongoUrl, err := connstring.ParseAndValidate(config.Config.MongoUri)
	if err != nil {
		err = &ConnectionError{
			errors.Wrap(err, "database: Failed to parse mongo uri"),
		}
		return
	}

	logrus.WithFields(logrus.Fields{
		"mongodb_hosts": mongoUrl.Hosts,
	}).Info("database: Connecting to MongoDB server")

	if mongoUrl.Database != "" {
		DefaultDatabase = mongoUrl.Database
	}

	opts := options.Client().ApplyURI(config.Config.MongoUri)
	opts.SetRetryReads(true)
	opts.SetRetryWrites(true)
	opts.WriteConcern = writeconcern.New(
		writeconcern.WMajority(),
		writeconcern.WTimeout(15*time.Second),
	)
	opts.ReadConcern = readconcern.Local()

	client, err := mongo.Connect(context.Background(), opts)
	if err != nil {
		err = &ConnectionError{
			errors.Wrap(err, "database: Connection error"),
		}
		return
	}

	setClient(client)

	err = ValidateDatabase()
	if err != nil {
		return
	}

	logrus.WithFields(logrus.Fields{
		"mongodb_hosts": mongoUrl.Hosts,
	}).Info("database: Connected to MongoDB server")

	err = addCollections()
	if err != nil {
		return
	}

	err = addIndexes()
	if err != nil {
		return
	}

	return
}

func ValidateDatabase() (err error) {
	db := GetDatabase()

	cursor, err := db.database.ListCollections(
		db, &bson.M{})
	if err != nil {
		err = ParseError(err)
		return
	}
	defer cursor.Close(db)

	for cursor.Next(db) {
		item := &struct {
			Name string `bson:"name"`
		}{}
		err = cursor.Decode(item)
		if err != nil {
			err = ParseError(err)
			return
		}

		if item.Name == "servers" {
			err = &errortypes.DatabaseError{
				errors.New("database: Cannot connect to pritunl database"),
			}
			return
		}
	}

	err = cursor.Err()
	if err != nil {
		err = ParseError(err)
		return
	}

	return
}

func getDatabase(ctx context.Context, client *mongo.Client) *Database {
	if client == nil {
		return nil
	}

	database := client.Database(DefaultDatabase)

	return &Database{
		ctx:      ctx,
		client:   client,
		database: database,
	}
}

func GetDatabase() *Database {
	return getDatabase(nil, getClient())
}

func GetDatabaseCtx(ctx context.Context) *Database {
	return getDatabase(ctx, getClient())
}

func addIndexes() (err error) {
	db := GetDatabase()
	defer db.Close()

	index := &Index{
		Collection: db.Users(),
		Keys: &bson.D{
			{"username", 1},
		},
	}
	err = index.Create()
	if err != nil {
		return
	}
	index = &Index{
		Collection: db.Users(),
		Keys: &bson.D{
			{"type", 1},
		},
	}
	err = index.Create()
	if err != nil {
		return
	}
	index = &Index{
		Collection: db.Users(),
		Keys: &bson.D{
			{"roles", 1},
		},
	}
	err = index.Create()
	if err != nil {
		return
	}
	index = &Index{
		Collection: db.Users(),
		Keys: &bson.D{
			{"token", 1},
		},
	}
	err = index.Create()
	if err != nil {
		return
	}

	index = &Index{
		Collection: db.Logs(),
		Keys: &bson.D{
			{"timestamp", 1},
		},
		Expire: 4320 * time.Hour,
	}
	err = index.Create()
	if err != nil {
		return
	}

	index = &Index{
		Collection: db.Audits(),
		Keys: &bson.D{
			{"user", 1},
		},
	}
	err = index.Create()
	if err != nil {
		return
	}

	index = &Index{
		Collection: db.Policies(),
		Keys: &bson.D{
			{"roles", 1},
		},
	}
	err = index.Create()
	if err != nil {
		return
	}

	index = &Index{
		Collection: db.CsrfTokens(),
		Keys: &bson.D{
			{"timestamp", 1},
		},
		Expire: 168 * time.Hour,
	}
	err = index.Create()
	if err != nil {
		return
	}

	index = &Index{
		Collection: db.SecondaryTokens(),
		Keys: &bson.D{
			{"timestamp", 1},
		},
		Expire: 3 * time.Minute,
	}
	err = index.Create()
	if err != nil {
		return
	}

	index = &Index{
		Collection: db.Nodes(),
		Keys: &bson.D{
			{"name", 1},
		},
	}
	err = index.Create()
	if err != nil {
		return
	}
	index = &Index{
		Collection: db.Nodes(),
		Keys: &bson.D{
			{"pools", 1},
		},
	}
	err = index.Create()
	if err != nil {
		return
	}

	index = &Index{
		Collection: db.NodePorts(),
		Keys: &bson.D{
			{"datacenter", 1},
			{"protocol", 1},
			{"port", 1},
		},
		Unique: true,
	}
	err = index.Create()
	if err != nil {
		return
	}
	index = &Index{
		Collection: db.NodePorts(),
		Keys: &bson.D{
			{"resource", 1},
		},
	}
	err = index.Create()
	if err != nil {
		return
	}

	index = &Index{
		Collection: db.Nonces(),
		Keys: &bson.D{
			{"timestamp", 1},
		},
		Expire: 24 * time.Hour,
	}
	err = index.Create()
	if err != nil {
		return
	}

	index = &Index{
		Collection: db.Rokeys(),
		Keys: &bson.D{
			{"type", 1},
			{"timeblock", 1},
		},
		Unique: true,
	}
	err = index.Create()
	if err != nil {
		return
	}
	index = &Index{
		Collection: db.Rokeys(),
		Keys: &bson.D{
			{"timestamp", 1},
		},
		Expire: 720 * time.Hour,
	}
	err = index.Create()
	if err != nil {
		return
	}

	index = &Index{
		Collection: db.Devices(),
		Keys: &bson.D{
			{"user", 1},
			{"mode", 1},
		},
	}
	err = index.Create()
	if err != nil {
		return
	}
	index = &Index{
		Collection: db.Devices(),
		Keys: &bson.D{
			{"provider", 1},
		},
	}
	err = index.Create()
	if err != nil {
		return
	}

	index = &Index{
		Collection: db.Alerts(),
		Keys: &bson.D{
			{"name", 1},
		},
	}
	err = index.Create()
	if err != nil {
		return
	}

	index = &Index{
		Collection: db.Alerts(),
		Keys: &bson.D{
			{"roles", 1},
		},
	}
	err = index.Create()
	if err != nil {
		return
	}

	index = &Index{
		Collection: db.AlertsEvent(),
		Keys: &bson.D{
			{"timestamp", 1},
		},
		Expire: 48 * time.Hour,
	}
	err = index.Create()
	if err != nil {
		return
	}

	index = &Index{
		Collection: db.AlertsEvent(),
		Keys: &bson.D{
			{"source", 1},
			{"resource", 1},
			{"timestamp", -1},
		},
	}
	err = index.Create()
	if err != nil {
		return
	}

	index = &Index{
		Collection: db.AlertsEventLock(),
		Keys: &bson.D{
			{"timestamp", 1},
		},
		Expire: 72 * time.Hour,
	}
	err = index.Create()
	if err != nil {
		return
	}

	index = &Index{
		Collection: db.Organizations(),
		Keys: &bson.D{
			{"name", 1},
		},
	}
	err = index.Create()
	if err != nil {
		return
	}
	index = &Index{
		Collection: db.Organizations(),
		Keys: &bson.D{
			{"roles", 1},
		},
	}
	err = index.Create()
	if err != nil {
		return
	}

	index = &Index{
		Collection: db.Images(),
		Keys: &bson.D{
			{"key", 1},
		},
	}
	err = index.Create()
	if err != nil {
		return
	}
	index = &Index{
		Collection: db.Images(),
		Keys: &bson.D{
			{"organization", 1},
			{"name", 1},
		},
	}
	err = index.Create()
	if err != nil {
		return
	}
	index = &Index{
		Collection: db.Images(),
		Keys: &bson.D{
			{"storage", 1},
			{"key", 1},
		},
		Unique: true,
	}
	err = index.Create()
	if err != nil {
		return
	}
	index = &Index{
		Collection: db.Images(),
		Keys: &bson.D{
			{"disk", 1},
		},
	}
	err = index.Create()
	if err != nil {
		return
	}

	index = &Index{
		Collection: db.LvmLock(),
		Keys: &bson.D{
			{"timestamp", 1},
		},
		Expire: 90 * time.Second,
	}
	err = index.Create()
	if err != nil {
		return
	}

	index = &Index{
		Collection: db.Disks(),
		Keys: &bson.D{
			{"instance", 1},
			{"index", 1},
		},
		Unique: true,
	}
	err = index.Create()
	if err != nil {
		return
	}
	index = &Index{
		Collection: db.Disks(),
		Keys: &bson.D{
			{"name", 1},
		},
	}
	err = index.Create()
	if err != nil {
		return
	}
	index = &Index{
		Collection: db.Disks(),
		Keys: &bson.D{
			{"organization", 1},
			{"name", 1},
		},
	}
	err = index.Create()
	if err != nil {
		return
	}
	index = &Index{
		Collection: db.Disks(),
		Keys: &bson.D{
			{"node", 1},
		},
	}
	err = index.Create()
	if err != nil {
		return
	}

	index = &Index{
		Collection: db.Domains(),
		Keys: &bson.D{
			{"last_update", -1},
		},
	}
	err = index.Create()
	if err != nil {
		return
	}

	index = &Index{
		Collection: db.DomainsRecords(),
		Keys: &bson.D{
			{"domain", 1},
		},
	}
	err = index.Create()
	if err != nil {
		return
	}

	index = &Index{
		Collection: db.DomainsRecords(),
		Keys: &bson.D{
			{"domain", 1},
			{"sub_domain", 1},
			{"type", 1},
			{"value", 1},
		},
		Unique: true,
	}
	err = index.Create()
	if err != nil {
		return
	}

	index = &Index{
		Collection: db.Datacenters(),
		Keys: &bson.D{
			{"organization", 1},
		},
	}
	err = index.Create()
	if err != nil {
		return
	}
	index = &Index{
		Collection: db.Datacenters(),
		Keys: &bson.D{
			{"match_organizations", 1},
		},
	}
	err = index.Create()
	if err != nil {
		return
	}

	index = &Index{
		Collection: db.BlocksIp(),
		Keys: &bson.D{
			{"block", 1},
			{"ip", 1},
		},
		Unique: true,
	}
	err = index.Create()
	if err != nil {
		return
	}
	index = &Index{
		Collection: db.BlocksIp(),
		Keys: &bson.D{
			{"instance", 1},
		},
	}
	err = index.Create()
	if err != nil {
		return
	}

	index = &Index{
		Collection: db.Vpcs(),
		Keys: &bson.D{
			{"name", 1},
		},
	}
	err = index.Create()
	if err != nil {
		return
	}
	index = &Index{
		Collection: db.Vpcs(),
		Keys: &bson.D{
			{"organization", 1},
			{"name", 1},
		},
	}
	err = index.Create()
	if err != nil {
		return
	}
	index = &Index{
		Collection: db.Vpcs(),
		Keys: &bson.D{
			{"vpc_id", 1},
		},
		Unique: true,
	}
	err = index.Create()
	if err != nil {
		return
	}
	index = &Index{
		Collection: db.Vpcs(),
		Keys: &bson.D{
			{"datacenter", 1},
		},
	}
	err = index.Create()
	if err != nil {
		return
	}

	index = &Index{
		Collection: db.VpcsIp(),
		Keys: &bson.D{
			{"vpc", 1},
			{"subnet", 1},
			{"ip", 1},
		},
		Unique: true,
	}
	err = index.Create()
	if err != nil {
		return
	}
	index = &Index{
		Collection: db.VpcsIp(),
		Keys: &bson.D{
			{"vpc", 1},
			{"instance", 1},
		},
	}
	err = index.Create()
	if err != nil {
		return
	}
	index = &Index{
		Collection: db.VpcsIp(),
		Keys: &bson.D{
			{"vpc", 1},
			{"subnet", 1},
		},
	}
	err = index.Create()
	if err != nil {
		return
	}

	index = &Index{
		Collection: db.Sessions(),
		Keys: &bson.D{
			{"user", 1},
		},
	}
	err = index.Create()
	if err != nil {
		return
	}

	index = &Index{
		Collection: db.Sessions(),
		Keys: &bson.D{
			{"timestamp", 1},
		},
		Expire: 4320 * time.Hour,
	}
	err = index.Create()
	if err != nil {
		return
	}

	index = &Index{
		Collection: db.Firewalls(),
		Keys: &bson.D{
			{"name", 1},
		},
	}
	err = index.Create()
	if err != nil {
		return
	}
	index = &Index{
		Collection: db.Firewalls(),
		Keys: &bson.D{
			{"organization", 1},
			{"name", 1},
		},
	}
	err = index.Create()
	if err != nil {
		return
	}
	index = &Index{
		Collection: db.Firewalls(),
		Keys: &bson.D{
			{"roles", 1},
		},
	}
	err = index.Create()
	if err != nil {
		return
	}
	index = &Index{
		Collection: db.Firewalls(),
		Keys: &bson.D{
			{"roles", 1},
			{"organization", 1},
		},
	}
	err = index.Create()
	if err != nil {
		return
	}

	index = &Index{
		Collection: db.Zones(),
		Keys: &bson.D{
			{"datacenter", 1},
		},
	}
	err = index.Create()
	if err != nil {
		return
	}

	index = &Index{
		Collection: db.Balancers(),
		Keys: &bson.D{
			{"datacenter", 1},
		},
	}
	err = index.Create()
	if err != nil {
		return
	}

	index = &Index{
		Collection: db.Authorities(),
		Keys: &bson.D{
			{"name", 1},
		},
	}
	err = index.Create()
	if err != nil {
		return
	}
	index = &Index{
		Collection: db.Authorities(),
		Keys: &bson.D{
			{"organization", 1},
			{"name", 1},
		},
	}
	err = index.Create()
	if err != nil {
		return
	}
	index = &Index{
		Collection: db.Authorities(),
		Keys: &bson.D{
			{"network_roles", 1},
		},
	}
	err = index.Create()
	if err != nil {
		return
	}
	index = &Index{
		Collection: db.Authorities(),
		Keys: &bson.D{
			{"organization", 1},
			{"network_roles", 1},
		},
	}
	err = index.Create()
	if err != nil {
		return
	}

	index = &Index{
		Collection: db.Instances(),
		Keys: &bson.D{
			{"node", 1},
		},
	}
	err = index.Create()
	if err != nil {
		return
	}
	index = &Index{
		Collection: db.Instances(),
		Keys: &bson.D{
			{"name", 1},
		},
	}
	err = index.Create()
	if err != nil {
		return
	}
	index = &Index{
		Collection: db.Instances(),
		Keys: &bson.D{
			{"organization", 1},
			{"name", 1},
		},
	}
	err = index.Create()
	if err != nil {
		return
	}
	index = &Index{
		Collection: db.Instances(),
		Keys: &bson.D{
			{"node", 1},
			{"vnc_display", 1},
		},
		Partial: &bson.M{
			"vnc_display": &bson.M{
				"$gt": 0,
			},
		},
		Unique: true,
	}
	err = index.Create()
	if err != nil {
		return
	}
	index = &Index{
		Collection: db.Instances(),
		Keys: &bson.D{
			{"node", 1},
			{"spice_port", 1},
		},
		Partial: &bson.M{
			"spice_port": &bson.M{
				"$gt": 0,
			},
		},
		Unique: true,
	}
	err = index.Create()
	if err != nil {
		return
	}
	index = &Index{
		Collection: db.Instances(),
		Keys: &bson.D{
			{"unix_id", 1},
		},
		Partial: &bson.M{
			"unix_id": &bson.M{
				"$gt": 0,
			},
		},
		Unique: true,
	}
	err = index.Create()
	if err != nil {
		return
	}
	index = &Index{
		Collection: db.Instances(),
		Keys: &bson.D{
			{"node_ports.node_port", 1},
		},
	}
	err = index.Create()
	if err != nil {
		return
	}

	index = &Index{
		Collection: db.Units(),
		Keys: &bson.D{
			{"pod", 1},
		},
	}
	err = index.Create()
	if err != nil {
		return
	}

	index = &Index{
		Collection: db.Tasks(),
		Keys: &bson.D{
			{"timestamp", 1},
		},
		Expire: 48 * time.Hour,
	}
	err = index.Create()
	if err != nil {
		return
	}

	index = &Index{
		Collection: db.Events(),
		Keys: &bson.D{
			{"channel", 1},
		},
	}
	err = index.Create()
	if err != nil {
		return
	}

	index = &Index{
		Collection: db.AcmeChallenges(),
		Keys: &bson.D{
			{"timestamp", 1},
		},
		Expire: 3 * time.Minute,
	}
	err = index.Create()
	if err != nil {
		return
	}

	index = &Index{
		Collection: db.Geo(),
		Keys: &bson.D{
			{"t", 1},
		},
		Expire: 360 * time.Hour,
	}
	err = index.Create()
	if err != nil {
		return
	}

	return
}

func addCollections() (err error) {
	db := GetDatabase()
	defer db.Close()

	cursor, err := db.database.ListCollections(
		db, &bson.M{})
	if err != nil {
		err = ParseError(err)
		return
	}
	defer cursor.Close(db)

	eventsExists := false
	isCapped := false

	for cursor.Next(db) {
		item := &struct {
			Name    string `bson:"name"`
			Options bson.M `bson:"options"`
		}{}
		err = cursor.Decode(item)
		if err != nil {
			err = ParseError(err)
			return
		}

		if item.Name == "events" {
			eventsExists = true
			if options, ok := item.Options["capped"]; ok {
				if cappedBool, ok := options.(bool); ok && cappedBool {
					isCapped = true
				}
			}
			break
		}
	}

	err = cursor.Err()
	if err != nil {
		err = ParseError(err)
		return
	}

	if eventsExists && !isCapped {
		logrus.WithFields(logrus.Fields{
			"collection": "events",
		}).Warning("database: Correcting events capped collection")

		err = db.database.Collection("events").Drop(context.Background())
		if err != nil {
			err = ParseError(err)
			return
		}
		eventsExists = false
	}

	if !eventsExists {
		err = db.database.RunCommand(
			context.Background(),
			bson.D{
				{"create", "events"},
				{"capped", true},
				{"max", 1000},
				{"size", 5242880},
			},
		).Err()
		if err != nil {
			err = ParseError(err)
			return
		}
	}

	return
}

func init() {
	module := requires.New("database")
	module.After("config")

	module.Handler = func() (err error) {
		for {
			e := Connect()
			if e != nil {
				logrus.WithFields(logrus.Fields{
					"error": e,
				}).Error("database: Connection error")
			} else {
				break
			}

			time.Sleep(constants.RetryDelay)
		}

		return
	}
}
