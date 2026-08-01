// Package subscription provides the subscription and billing domain: the
// Subscription model, its repository, Service.CreateSubscription (charge via
// the payment subpackage, persist only on success), and a Resolver adapter
// feeding jwt.Refresher's has_subscription re-check. It has no HTTP handler
// or routes of its own — registration and login reach it only through
// internal/auth's composition (cmd/api/auth.go).
package subscription
