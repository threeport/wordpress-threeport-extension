// generated by 'threeport-sdk gen' but will not be regenerated - intended for modification

package v0

import (
	"fmt"
	tpapi_v0 "github.com/threeport/threeport/pkg/api/v0"
	util "github.com/threeport/threeport/pkg/util/v0"
	api_v0 "github.com/threeport/wordpress-threeport-extension/pkg/api/v0"
	client_v0 "github.com/threeport/wordpress-threeport-extension/pkg/client/v0"
	"net/http"
)

// WordpressConfig contains the config for a wordpress which is an abstraction
// of a wordpress definition and wordpress instance.
type WordpressConfig struct {
	Wordpress WordpressValues `yaml:"Wordpress"`
}

// WordpressValues contains the attributes needed to manage a wordpress
// definition and wordpress instance with a single operation.
type WordpressValues struct {
	Name string `yaml:"Name"`
}

// Create creates a wordpress definition and instance in the Threeport API.
func (w *WordpressValues) Create(
	apiClient *http.Client,
	apiEndpoint string,
) (*api_v0.WordpressDefinition, *api_v0.WordpressInstance, error) {
	// get operations
	operations, createdWordpressDefinition, createdWordpressInstance := w.GetOperations(
		apiClient,
		apiEndpoint,
	)

	// execute create operations
	if err := operations.Create(); err != nil {
		return nil, nil, fmt.Errorf(
			"failed to execute create operations for wordpress defined instance with name %s: %w",
			w.Name,
			err,
		)
	}

	return createdWordpressDefinition, createdWordpressInstance, nil
}

// Delete deletes a wordpress definition and instance from the Threeport API.
func (w *WordpressValues) Delete(
	apiClient *http.Client,
	apiEndpoint string,
) (*api_v0.WordpressDefinition, *api_v0.WordpressInstance, error) {
	// get operations
	operations, _, _ := w.GetOperations(
		apiClient,
		apiEndpoint,
	)

	// execute delete operations
	if err := operations.Delete(); err != nil {
		return nil, nil, fmt.Errorf(
			"failed to execute delete operations for wordpress defined instance with name %s: %w",
			w.Name,
			err,
		)
	}

	return nil, nil, nil
}

// GetOperations returns a slice of operations used to create or delete a
// wordpress defined instance.
func (w *WordpressValues) GetOperations(
	apiClient *http.Client,
	apiEndpoint string,
) (*util.Operations, *api_v0.WordpressDefinition, *api_v0.WordpressInstance) {
	var err error
	var createdWordpressDefinition api_v0.WordpressDefinition
	var createdWordpressInstance api_v0.WordpressInstance

	operations := util.Operations{}

	// add wordpress definition operation
	wordpressDefinitionValues := WordpressDefinitionValues{
		Name: w.Name,
	}
	operations.AppendOperation(util.Operation{
		Create: func() error {
			wordpressDefinition, err := wordpressDefinitionValues.Create(apiClient, apiEndpoint)
			if err != nil {
				return fmt.Errorf("failed to create wordpress definition with name %s: %w", w.Name, err)
			}
			createdWordpressDefinition = *wordpressDefinition
			return nil
		},
		Delete: func() error {
			_, err = wordpressDefinitionValues.Delete(apiClient, apiEndpoint)
			if err != nil {
				return fmt.Errorf("failed to delete wordpress definition with name %s: %w", w.Name, err)
			}
			return nil
		},
		Name: "wordpress definition",
	})

	// add wordpress instance operation
	wordpressInstanceValues := WordpressInstanceValues{
		Name: w.Name,
	}
	operations.AppendOperation(util.Operation{
		Create: func() error {
			wordpressInstance, err := wordpressInstanceValues.Create(apiClient, apiEndpoint)
			if err != nil {
				return fmt.Errorf("failed to create wordpress instance with name %s: %w", w.Name, err)
			}
			createdWordpressInstance = *wordpressInstance
			return nil
		},
		Delete: func() error {
			_, err = wordpressInstanceValues.Delete(apiClient, apiEndpoint)
			if err != nil {
				return fmt.Errorf("failed to delete wordpress instance with name %s: %w", w.Name, err)
			}
			return nil
		},
		Name: "wordpress instance",
	})

	return &operations, &createdWordpressDefinition, &createdWordpressInstance
}

// WordpressDefinitionConfig contains the config for a wordpress definition.
type WordpressDefinitionConfig struct {
	WordpressDefinition WordpressDefinitionValues `yaml:"WordpressDefinition"`
}

// WordpressDefinitionValues contains the attributes for the wordpress definition
// config abstraction.
type WordpressDefinitionValues struct {
	Name string `yaml:"Name"`
}

// Create creates a wordpress definition in the Threeport API.
func (w *WordpressDefinitionValues) Create(
	apiClient *http.Client,
	apiEndpoint string,
) (*api_v0.WordpressDefinition, error) {
	// validate config
	// TODO

	// construct wordpress definition object
	wordpressDefinition := api_v0.WordpressDefinition{
		Definition: tpapi_v0.Definition{
			Name: &w.Name,
		},
	}

	// create wordpress definition
	createdWordpressDefinition, err := client_v0.CreateWordpressDefinition(
		apiClient,
		apiEndpoint,
		&wordpressDefinition,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create wordpress definition in threeport API: %w", err)
	}

	return createdWordpressDefinition, nil
}

// Delete deletes a wordpress definition from the Threeport API.
func (w *WordpressDefinitionValues) Delete(
	apiClient *http.Client,
	apiEndpoint string,
) (*api_v0.WordpressDefinition, error) {
	// get wordpress definition by name
	wordpressDefinition, err := client_v0.GetWordpressDefinitionByName(
		apiClient,
		apiEndpoint,
		w.Name,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to find wordpress definition with name %s: %w", w.Name, err)
	}

	// delete wordpress definition
	deletedWordpressDefinition, err := client_v0.DeleteWordpressDefinition(
		apiClient,
		apiEndpoint,
		*wordpressDefinition.ID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to delete wordpress definition from Threeport API: %w", err)
	}

	return deletedWordpressDefinition, nil
}

// WordpressInstanceConfig contains the config for a wordpress instance.
type WordpressInstanceConfig struct {
	WordpressInstance WordpressInstanceValues `yaml:"WordpressInstance"`
}

// WordpressInstanceValues contains the attributes for the wordpress instance
// config abstraction.
type WordpressInstanceValues struct {
	Name string `yaml:"Name"`
}

// Create creates a wordpress instance in the Threeport API.
func (w *WordpressInstanceValues) Create(
	apiClient *http.Client,
	apiEndpoint string,
) (*api_v0.WordpressInstance, error) {
	// validate config
	// TODO

	// construct wordpress instance object
	wordpressInstance := api_v0.WordpressInstance{
		Instance: tpapi_v0.Instance{
			Name: &w.Name,
		},
	}

	// create wordpress instance
	createdWordpressInstance, err := client_v0.CreateWordpressInstance(
		apiClient,
		apiEndpoint,
		&wordpressInstance,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create wordpress instance in threeport API: %w", err)
	}

	return createdWordpressInstance, nil
}

// Delete deletes a wordpress instance from the Threeport API.
func (w *WordpressInstanceValues) Delete(
	apiClient *http.Client,
	apiEndpoint string,
) (*api_v0.WordpressInstance, error) {
	// get wordpress instance by name
	wordpressInstance, err := client_v0.GetWordpressInstanceByName(
		apiClient,
		apiEndpoint,
		w.Name,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to find wordpress instance with name %s: %w", w.Name, err)
	}

	// delete wordpress instance
	deletedWordpressInstance, err := client_v0.DeleteWordpressInstance(
		apiClient,
		apiEndpoint,
		*wordpressInstance.ID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to delete wordpress instance from Threeport API: %w", err)
	}

	return deletedWordpressInstance, nil
}