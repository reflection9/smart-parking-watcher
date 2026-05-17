locals {
  common_labels = {
    "app.kubernetes.io/part-of"    = "smart-parking-watcher"
    "app.kubernetes.io/managed-by" = "terraform"
  }

  namespaces = {
    smart_parking = var.smart_parking_namespace
    argocd        = var.argocd_namespace
    kafka         = var.kafka_namespace
    observability = var.observability_namespace
  }
}

resource "kubernetes_namespace_v1" "namespaces" {
  for_each = local.namespaces

  metadata {
    name = each.value

    labels = merge(
      local.common_labels,
      {
        "app.kubernetes.io/name" = each.value
      }
    )
  }

  wait_for_default_service_account = true
}

resource "kubernetes_service_account_v1" "smart_parking_runtime" {
  metadata {
    name      = "smart-parking-runtime"
    namespace = kubernetes_namespace_v1.namespaces["smart_parking"].metadata[0].name

    labels = merge(
      local.common_labels,
      {
        "app.kubernetes.io/name"      = "smart-parking-runtime"
        "app.kubernetes.io/component" = "runtime"
      }
    )
  }

  automount_service_account_token = false
}

resource "kubernetes_service_account_v1" "smart_parking_migrations" {
  metadata {
    name      = "smart-parking-migrations"
    namespace = kubernetes_namespace_v1.namespaces["smart_parking"].metadata[0].name

    labels = merge(
      local.common_labels,
      {
        "app.kubernetes.io/name"      = "smart-parking-migrations"
        "app.kubernetes.io/component" = "database"
      }
    )
  }

  automount_service_account_token = false
}

resource "kubernetes_service_account_v1" "argocd_bootstrap" {
  metadata {
    name      = "argocd-bootstrap"
    namespace = kubernetes_namespace_v1.namespaces["argocd"].metadata[0].name

    labels = merge(
      local.common_labels,
      {
        "app.kubernetes.io/name"      = "argocd-bootstrap"
        "app.kubernetes.io/component" = "gitops"
      }
    )
  }
}

resource "kubernetes_secret_v1" "smart_parking_app_secret" {
  metadata {
    name      = "smart-parking-app-secret"
    namespace = kubernetes_namespace_v1.namespaces["smart_parking"].metadata[0].name

    labels = merge(
      local.common_labels,
      {
        "app.kubernetes.io/name"      = "smart-parking-app-secret"
        "app.kubernetes.io/component" = "configuration"
      }
    )
  }

  type = "Opaque"

  data = {
    DB_PASSWORD      = var.app_secret_data.DB_PASSWORD
    MINIO_ACCESS_KEY = var.app_secret_data.MINIO_ACCESS_KEY
    MINIO_SECRET_KEY = var.app_secret_data.MINIO_SECRET_KEY
    SMTP_USERNAME    = var.app_secret_data.SMTP_USERNAME
    SMTP_PASSWORD    = var.app_secret_data.SMTP_PASSWORD
  }
}
