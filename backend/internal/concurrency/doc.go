// Package concurrency provides optimistic locking: version/timestamp-based
// conflict detection for concurrent writes to trips and activities (the
// versioned itinerary entity — see migrations/010_create_trips_table.sql and
// migrations/015_create_activities_table.sql's "Naming correction" notes;
// there is no separate itinerary_items table). This package itself is still
// scaffolding for Spec 004 (Security & Authentication/Authorization Model,
// see specs/004-security-auth-model/): the version columns already exist on
// trips/activities, but a shared conflict-detection implementation lands in
// a later Spec 004 issue.
package concurrency
