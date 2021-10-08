variable "host" {
  type = string
}

variable "port" {
  type = number
}

variable "name" {
  type = string
}

variable "namespace" {
  type = string
}

variable "identifier" {
  type = string
}

variable "allowed_users" {
  type = set(string)
}

variable "scheme" {
  type = string
}

variable "db_name" {
  type = string
}
