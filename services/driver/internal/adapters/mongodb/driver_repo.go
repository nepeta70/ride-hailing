package mongodb

import (
	"context"
	"time"

	"github.com/google/uuid"
	pkgMongo "github.com/nepeta70/ride-hailing/internal/pkg/adapters/mongodb"
	"github.com/nepeta70/ride-hailing/internal/pkg/errors"
	pkgPorts "github.com/nepeta70/ride-hailing/internal/pkg/ports"
	"github.com/nepeta70/ride-hailing/services/driver/internal/config"
	"github.com/nepeta70/ride-hailing/services/driver/internal/core/domain"
	"github.com/nepeta70/ride-hailing/services/driver/internal/ports"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type DriverRepository struct {
	db         *mongo.Database
	collection *mongo.Collection
	telemetry  pkgPorts.TelemetryProvider
}

// NewDriverRepository initializes the connection and sets up indexes
func NewDriverRepository(ctx context.Context, cfg *config.Config, mongoAdapter *pkgMongo.MongoAdapter, telemetry pkgPorts.TelemetryProvider) (ports.DriverRepository, error) {
	db := mongoAdapter.Client.Database(cfg.Mongo.Database)
	col := db.Collection("drivers")

	repo := &DriverRepository{
		db:         db,
		collection: col,
		telemetry:  telemetry,
	}

	if err := repo.ensureSchema(ctx); err != nil {
		return nil, errors.NewTransientErrorf("failed to initialize mongodb schema: %w", err)
	}
	return repo, nil
}

func (r *DriverRepository) GetDriver(ctx context.Context, userID uuid.UUID) (*domain.Driver, error) {
	var doc DriverDoc
	err := r.collection.FindOne(ctx, bson.M{"user_id": userID.String()}).Decode(&doc)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			// Return your custom "NotFound" error that the CB ignores
			return nil, errors.NewErrNotFound("driver profile not found")
		}
		return nil, err
	}
	return doc.ToDomain(), nil
}

// AddDriver inserts a new driver document into the collection
func (r *DriverRepository) AddDriver(ctx context.Context, driver *domain.Driver) error {
	now := time.Now().UTC()

	driverDoc := FromDomain(driver)
	driver.CreatedAt = now
	driver.UpdatedAt = now

	// Ensure the ID is generated if not provided
	if driverDoc.ID.IsZero() {
		driverDoc.ID = primitive.NewObjectID()
	}

	_, err := r.collection.InsertOne(ctx, driverDoc)
	if err != nil {
		// Handle duplicate key error (if user_id already exists)
		if mongo.IsDuplicateKeyError(err) {
			return errors.NewBusinessError("driver profile already exists for this user")
		}
		return err
	}

	return nil
}

// UpdateDriver updates an existing driver's profile based on their UserID
func (r *DriverRepository) UpdateDriver(ctx context.Context, driver *domain.Driver) error {
	now := time.Now().UTC()

	// Define the filter to find the specific driver
	filter := bson.M{"user_id": driver.UserID}

	// Define the update operation using $set
	update := bson.M{
		"$set": bson.M{
			"license_number": driver.LicenseNumber,
			"license_expiry": driver.LicenseExpiry,
			"vehicle":        driver.Vehicle,
			"updated_at":     now,
		},
	}

	// Execute the update
	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	// If no document matched the user_id, it means the driver doesn't exist
	if result.MatchedCount == 0 {
		return errors.NewErrNotFound("cannot update: driver profile not found")
	}

	return nil
}

func (r *DriverRepository) ensureSchema(ctx context.Context) error {
	_, err := r.collection.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "user_id", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys:    bson.D{{Key: "vehicle.license_plate", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
	})
	return err
}

var _ ports.DriverReadRepository = (*DriverRepository)(nil)
var _ ports.DriverWriteRepository = (*DriverRepository)(nil)
