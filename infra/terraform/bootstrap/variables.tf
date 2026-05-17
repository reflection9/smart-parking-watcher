variable "kubeconfig_path" {
  description = "Path to the kubeconfig file used by the Kubernetes provider."
  type        = string
  default     = "~/.kube/config"
}

variable "smart_parking_namespace" {
  description = "Namespace for SmartParkingWatcher application workloads."
  type        = string
  default     = "smart-parking"
}

variable "argocd_namespace" {
  description = "Namespace where Argo CD is installed."
  type        = string
  default     = "argocd"
}

variable "kafka_namespace" {
  description = "Namespace for Strimzi and Kafka resources."
  type        = string
  default     = "kafka"
}

variable "observability_namespace" {
  description = "Namespace reserved for observability stack resources."
  type        = string
  default     = "observability"
}

variable "app_secret_data" {
  description = "Base secret values for SmartParkingWatcher workloads."
  type = object({
    DB_PASSWORD      = string
    MINIO_ACCESS_KEY = string
    MINIO_SECRET_KEY = string
    SMTP_USERNAME    = string
    SMTP_PASSWORD    = string
  })
  sensitive = true
}
