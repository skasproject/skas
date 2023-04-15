package datawatcher

import "context"

type ParserFunc func(data string) (interface{}, error)

type DataWatcher interface {
	// Get will always return a valid value.
	// This will imply a first fetch of data has to be performed by the New() (Which may return an error if data is invalid
	Get() interface{}
	// Run is intended to be called by the github.com/pior/runnable package.
	// It must be blocking and end on <-ctx.Done()
	// Typically, it just call Start(ctx)
	Run(ctx context.Context) error
	// Start is intended to be called by the kubebuilder manager package
	// It must be blocking and end on <-ctx.Done()
	Start(ctx context.Context) error
}
