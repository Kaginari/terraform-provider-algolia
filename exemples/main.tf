terraform {
  required_version = ">= 0.13"

  required_providers {
    algolia = {
      source  = "Kaginari/algolia"
      version = "0.0.1"
    }
  }
}

provider "algolia" {
  application_id = "1JZFLS5GAW"
  api_key        = "154370f49d957ce17fa6a5d5fa1c1cd0"
}

resource "algolia_index" "index" {
  name = "test_index"
}

resource "algolia_api_key" "example" {
  acl         = ["search"]
  description = "example"
  indexes     = [algolia_index.index.name]
}



output "api_key" {
  value = algolia_api_key.example.key
}

output "hits_per_page" {
  value = algolia_index.index.name
}