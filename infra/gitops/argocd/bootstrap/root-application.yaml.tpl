apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: smart-parking-root
  namespace: argocd
spec:
  project: default
  source:
    repoURL: __REPO_URL__
    targetRevision: __TARGET_REVISION__
    path: infra/gitops/apps
    directory:
      recurse: true
  destination:
    server: https://kubernetes.default.svc
    namespace: argocd
  syncPolicy:
    automated:
      prune: true
      selfHeal: true
