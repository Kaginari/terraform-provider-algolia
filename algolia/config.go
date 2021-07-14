package algolia

import (
	"github.com/algolia/algoliasearch-client-go/v3/algolia/search"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

type apiClient struct {
	algolia *search.Client
}

func setValues(d *schema.ResourceData, values map[string]interface{}) error {
	for k, v := range values {
		if err := d.Set(k, v); err != nil {
			return err
		}
	}

	return nil
}
