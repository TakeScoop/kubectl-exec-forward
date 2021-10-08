resource "kubernetes_deployment" "this" {
  metadata {
    name      = var.name
    namespace = var.namespace
    labels = {
      app = var.name
    }
  }

  spec {
    replicas = 1

    selector {
      match_labels = {
        app = var.name
      }
    }

    template {
      metadata {
        labels = {
          app = var.name
        }
      }

      spec {
        container {
          image = "alpine/socat"
          name  = var.name
          command = [
            "socat",
            "tcp-listen:${var.port},fork,reuseaddr",
            "tcp-connect:${var.host}:${var.port}",
          ]
          port {
            name           = "forwarded"
            container_port = var.port
          }
        }
      }
    }
  }
}

resource "kubernetes_service" "this" {
  metadata {
    name      = var.name
    namespace = var.namespace
    annotations = {
      "aws-con.service.kubernetes.io/defaults" : jsonencode({
        username  = "iam_read",
        localport = "8080"
      })
      "aws-con.service.kubernetes.io/pre-commands" = jsonencode([
        {
          id      = "check"
          command = jsonencode(split(" ", "aws rds describe-db-instances --db-instance-identifier ${var.identifier}"))
        },
        {
          id      = "token"
          command = jsonencode(split(" ", "aws rds generate-db-auth-token --host ${var.host} --port ${var.port} --username {{.Config.username}}"))
        }
      ])
      "aws-con.service.kubernetes.io/post-commands" = jsonencode([
        {
          id      = "open"
          command = "[\"open\", \"${var.scheme}://{{ .Config.username }}:{{ urlquery (trim .Pre.token) }}@localhost:{{ .Config.localport }}/${var.db_name}\"]"
        },
      ])
    }
  }

  spec {
    selector = {
      app = kubernetes_deployment.this.metadata[0].labels.app
    }
    port {
      port        = var.port
      target_port = var.port
    }
  }
}
