package main

import (
	"github.com/Kaginari/terraform-provider-algolia/algolia"
	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: algolia.Provider,
	})
}
