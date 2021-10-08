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
      "aws-con.service.kubernetes.io/type" = "rds-iam"
      "aws-con.service.kubernetes.io/meta" = jsonencode({
        identifier    = "scoop-dev1-harbormaster-db"
        host          = "scoop-dev1-harbormaster-db.cqzu6spk5x9z.us-west-2.rds.amazonaws.com"
        port          = 5432
        db_name       = "harbormasterDb"
        protocol      = "postgresql"
        allowed_users = ["iam_read", "iam_read_write"]
      })
    }
  }
  spec {
    selector = {
      app = kubernetes_deployment.this.metadata[0].labels.app
    }
    port {
      name        = "forwarded"
      port        = var.port
      target_port = var.port
    }
  }
}
