module "forwarder" {
  source = "./modules/forwarder"

  name      = "harbormaster-db-proxy"
  namespace = "release"
  host      = "scoop-dev1-harbormaster-db.cqzu6spk5x9z.us-west-2.rds.amazonaws.com"
  port      = 5432
}
