package apperrors

import "errors"

var (
	// ErrNoSubscription is a sentinel error for subscription issues
	ErrNoSubscription = errors.New("no active subscription")
	// ErrMaxConfigsReached is a sentinel error for hitting config limits
	ErrMaxConfigsReached = errors.New("maximum number of configs reached")
	// ErrNotFound is a sentinel error for when a resource is not found
	ErrNotFound = errors.New("resource not found")
	// ErrUnauthorized is a sentinel error for when a user lacks permissions
	ErrUnauthorized = errors.New("unauthorized access")
)

// type ConfigNotFound struct {
// }

// func (e ConfigNotFound) Error() string {
// 	return "config not found"
// }

// func (e ConfigNotFound) Is(target error) bool {
// 	var configNotFound ConfigNotFound
// 	return errors.As(target, &configNotFound)
// }

// type MaxConfigLimitReached struct {
// }

// func (e MaxConfigLimitReached) Error() string {
// 	return "Cannot add more than 5 config"
// }

// func (e MaxConfigLimitReached) Is(target error) bool {
// 	var maxConfigLimitReached MaxConfigLimitReached
// 	ok := errors.As(target, &maxConfigLimitReached)
// 	return ok
// }

// type NoSubscriptionError struct {
// }

// func (e NoSubscriptionError) Error() string {
// 	return "no subscription"
// }

// func (e NoSubscriptionError) Is(target error) bool {
// 	var noSubscriptionError NoSubscriptionError
// 	ok := errors.As(target, &noSubscriptionError)
// 	return ok
// }
