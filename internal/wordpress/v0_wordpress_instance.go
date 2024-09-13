// generated by 'threeport-sdk gen' but will not be regenerated - intended for modification

package wordpress

import (
	logr "github.com/go-logr/logr"
	controller "github.com/threeport/threeport/pkg/controller/v0"
	v0 "github.com/threeport/wordpress-threeport-extension/pkg/api/v0"
)

// v0WordpressInstanceCreated performs reconciliation when a v0 WordpressInstance
// has been created.
func v0WordpressInstanceCreated(
	r *controller.Reconciler,
	wordpressInstance *v0.WordpressInstance,
	log *logr.Logger,
) (int64, error) {
	return 0, nil
}

// v0WordpressInstanceUpdated performs reconciliation when a v0 WordpressInstance
// has been updated.
func v0WordpressInstanceUpdated(
	r *controller.Reconciler,
	wordpressInstance *v0.WordpressInstance,
	log *logr.Logger,
) (int64, error) {
	return 0, nil
}

// v0WordpressInstanceDeleted performs reconciliation when a v0 WordpressInstance
// has been deleted.
func v0WordpressInstanceDeleted(
	r *controller.Reconciler,
	wordpressInstance *v0.WordpressInstance,
	log *logr.Logger,
) (int64, error) {
	return 0, nil
}
