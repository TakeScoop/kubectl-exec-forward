module "forwarder" {
  source = "./modules/forwarder"

  name          = "harbormaster-db-proxy"
  namespace     = "release"
  host          = "scoop-dev1-harbormaster-db.cqzu6spk5x9z.us-west-2.rds.amazonaws.com"
  port          = 5432
  allowed_users = ["iam_read", "iam_read_write"]
  identifier    = "scoop-dev1-harbormaster-db"
  scheme        = "postgres"
  db_name       = "harbormasterDb"
}
