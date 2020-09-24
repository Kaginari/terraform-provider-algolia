package main

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"
	"github.com/Kaginari/terraform-provider-algolia/algolia"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: algolia.Provider,
	})
}