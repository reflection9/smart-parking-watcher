output "namespace_names" {
  description = "Namespaces created by the Terraform bootstrap."
  value = {
    for key, namespace in kubernetes_namespace_v1.namespaces :
    key => namespace.metadata[0].name
  }
}

output "service_accounts" {
  description = "Service accounts created for bootstrap platform concerns."
  value = {
    smart_parking_runtime    = kubernetes_service_account_v1.smart_parking_runtime.metadata[0].name
    smart_parking_migrations = kubernetes_service_account_v1.smart_parking_migrations.metadata[0].name
    argocd_bootstrap         = kubernetes_service_account_v1.argocd_bootstrap.metadata[0].name
  }
}

output "smart_parking_app_secret_name" {
  description = "Name of the base application secret."
  value       = kubernetes_secret_v1.smart_parking_app_secret.metadata[0].name
}
