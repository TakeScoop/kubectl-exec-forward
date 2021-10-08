provider "kubernetes" {
  host                   = data.terraform_remote_state.kubernetes.outputs.apiserver_url
  cluster_ca_certificate = data.terraform_remote_state.kubernetes.outputs.terraform_service_account.ca
  token                  = data.terraform_remote_state.kubernetes.outputs.terraform_service_account.token
}
