db = db.getSiblingDB('driver_db');

db.createCollection('drivers');

db.drivers.createIndex({ "user_id": 1 }, { unique: true });

print("Driver DB initialized with unique user_id index.");