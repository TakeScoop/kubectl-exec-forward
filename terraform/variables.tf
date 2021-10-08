variable "host" {
  type    = string
  default = "scoop-dev1-harbormaster-db.cqzu6spk5x9z.us-west-2.rds.amazonaws.com"
}

variable "port" {
  type    = number
  default = 5432
}

variable "name" {
  type    = string
  default = "harbormaster-db-proxy"
}

variable "namespace" {
  type    = string
  default = "release"
}

variable "identifier" {
  type    = string
  default = "scoop-dev1-harbormaster-db"
}

variable "scheme" {
  type    = string
  default = "postgres"
}

variable "db_name" {
  type    = string
  default = "harbormasterDb"
}
