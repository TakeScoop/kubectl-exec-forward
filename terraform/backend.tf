data "terraform_remote_state" "kubernetes" {
  backend = "remote"

  config = {
    hostname     = "terraform.takescoop.com"
    organization = "takescoop"

    workspaces = {
      name = "kubernetes-staging"
    }
  }
}
