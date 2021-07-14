terraform {
  required_version = ">= 0.13"

  required_providers {
    algolia = {
      source = "Kaginari/algolia"
      version = "9.9.9"
    }

  }
}
provider "algolia" {

}
resource "algolia_index" "example_replica" {
  name = "example_replica"
}
resource "algolia_index" "example" {
  depends_on = [algolia_index.example_replica]
  name     = "example"
  replicas = [algolia_index.example_replica.name]
}
resource "algolia_api_key" "example" {
  depends_on = [algolia_index.example]
  acl         = ["search"]
  description = "example"
  indexes     = ["example"]
}
resource "algolia_index_rule" "rule" {
  index = algolia_index.example.name

  name = "rule-a-id"
  enabled = true
  apply_to_replicas = true

  consequence_params = "release_date >= 156849840"

  condition {
    pattern = "smartphone"
    anchoring = "startsWith"
    alternatives = true
  }


}


