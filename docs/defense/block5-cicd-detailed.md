# Блок 5. CI/CD И Helm Delivery

## Что Требовалось По ТЗ

В пятом блоке требовалось:

- поднять локальный CI runner
- собрать pipeline build/push/deploy
- подготовить Helm chart’ы минимум для `3+` микросервисов

## Что Реализовано

Блок сделан вокруг следующей схемы:

- `GitHub Actions self-hosted runner`
- `Kaniko` для сборки образов
- локальный registry `spw-registry:5000`
- `Helm` chart приложения
- `Argo CD` refresh после обновления chart values

Ключевые файлы:

- [.github/workflows/block5-local-cicd.yml](C:/Users/Mikolgi/GolandProjects/smart-parking-watcher/.github/workflows/block5-local-cicd.yml)
- [helm/smart-parking](C:/Users/Mikolgi/GolandProjects/smart-parking-watcher/helm/smart-parking)
- [scripts/ci/build-push-kaniko.ps1](C:/Users/Mikolgi/GolandProjects/smart-parking-watcher/scripts/ci/build-push-kaniko.ps1)
- [scripts/ci/update-helm-image-tag.ps1](C:/Users/Mikolgi/GolandProjects/smart-parking-watcher/scripts/ci/update-helm-image-tag.ps1)
- [scripts/ci/refresh-argocd-app.ps1](C:/Users/Mikolgi/GolandProjects/smart-parking-watcher/scripts/ci/refresh-argocd-app.ps1)
- [infra/gitops/apps/smart-parking-services.yaml](C:/Users/Mikolgi/GolandProjects/smart-parking-watcher/infra/gitops/apps/smart-parking-services.yaml)
- [infra/docker/go-service.Dockerfile](C:/Users/Mikolgi/GolandProjects/smart-parking-watcher/infra/docker/go-service.Dockerfile)

## Почему Вообще Этот Блок Сложнее

Этот блок сложнее остальных не потому, что `Helm` трудный, а потому что он соединяет сразу несколько контуров:

- Git source
- CI runner
- image build
- registry
- manifest packaging
- GitOps deployment

Если хотя бы один из этих слоев не договорился с другими, pipeline становится “формально написанным, но не рабочим”.

## Как Работает Сборка Образов

### Базовый Dockerfile

У проекта один общий Dockerfile:

- [infra/docker/go-service.Dockerfile](C:/Users/Mikolgi/GolandProjects/smart-parking-watcher/infra/docker/go-service.Dockerfile)

Он:

- берет `SERVICE_NAME` как build arg
- копирует shared observability module
- скачивает зависимости нужного сервиса
- собирает бинарник
- кладет его в легкий runtime image

Это хороший паттерн, потому что:

- не нужно держать шесть почти одинаковых Dockerfile
- build logic централизован
- CI проще поддерживать

### Почему `Kaniko`

`Kaniko` выбран для CI build-пути, потому что:

- это понятный CI-friendly image builder
- он не завязан на “docker build на хосте” как единственный путь
- это нормальный аргумент на защите, когда спрашивают про безопасную container build модель

## Почему Registry Не `localhost:5001`

Это один из самых важных каверзных вопросов.

Снаружи хоста registry проброшен как:

- `localhost:5001`

Но внутри контейнеров и `k3d` нод правильный адрес:

- `spw-registry:5000`

Причина:

- внутри CI-контейнера `localhost` означает сам контейнер
- `Kaniko` должен пушить не в себя, а в registry container в Docker network
- `k3d` уже знает registry по имени `spw-registry:5000` через `registries.yaml`

То есть:

- для человека на хосте удобно `localhost:5001`
- для CI и Kubernetes runtime корректно `spw-registry:5000`

## Как Работает Helm Layer

### Что Упаковано В Chart

Chart:

- собирает `smart-parking-app-config`
- собирает `smart-parking-app-secret`
- собирает `gateway-nginx-config`
- описывает backend services и deployments
- описывает `gateway`

То есть закрывает требование “минимум 3 микросервиса” с запасом.

### Почему Helm Вообще Нужен

Потому что иначе пришлось бы:

- обновлять image tag в куче raw YAML
- дублировать env, port, initContainer, resource patterns
- поддерживать сервисы тяжелее

С `Helm` мы переносим это в values-driven модель:

- меняем image tag в одном месте
- chart шаблонизирует сервисы
- Argo CD потом рендерит и применяет release

## Как Устроен Pipeline

Flow pipeline такой:

1. Workflow стартует на `self-hosted runner`
2. runner checkout’ит репозиторий
3. вычисляется новый image tag, обычно по SHA
4. `Kaniko` собирает образы сервисов
5. образы пушатся в `spw-registry:5000`
6. script обновляет `global.image.tag` в `helm/smart-parking/values.yaml`
7. workflow коммитит новое значение tag обратно в ветку
8. `Argo CD` app `smart-parking-services` получает refresh
9. кластер синхронизирует новый release

То есть у нас deployment идет не прямым `kubectl apply`, а через GitOps-путь.

## Почему `Argo CD` Не Заменен Прямым Helm Upgrade Из CI

Потому что основной deployment pattern проекта уже GitOps-based.

Если бы CI сам делал `helm upgrade --install`, то:

- Git перестал бы быть единственным источником истины
- кластерное состояние могло бы расходиться с репозиторием

Текущая модель правильнее:

- CI обновляет Git-managed values
- Argo CD подтягивает это изменение как desired state

## Что Изменилось В Argo App

Раньше `smart-parking-services` смотрел в raw manifests.

Теперь он смотрит в:

- `helm/smart-parking`

Это значит, что application delivery стал chart-based, а не file-based.

## Почему Нужен Self-Hosted Runner

Потому что pipeline зависит от локальной среды:

- локальный `k3d`
- локальный registry
- локальный kubeconfig
- доступ к Docker network `k3d-spw-local`

GitHub-hosted runner снаружи этого просто не увидит.

## Коварные Вопросы И Ответы

### Почему workflow коммитит values обратно в ветку

Потому что Argo CD должен увидеть новую desired state ревизию в Git. Если tag меняется только локально во время job, GitOps не узнает о новой версии.

### Почему не использовать внешний registry

Можно, но для локальной лабораторной среды уже есть встроенный локальный registry у `k3d`, и это уменьшает количество внешних зависимостей.

### Helm chart покрывает только сервисы, а не всю платформу. Это нормально

Да. Для этого блока достаточно закрыть delivery микросервисов и их конфигурации. Platform-level части уже живут отдельными блоками и отдельными Argo applications.

### Почему не GitLab Runner

Потому что текущий репозиторий и workflow проще и естественнее закрыть через `GitHub Actions self-hosted`.

## Честные Ограничения

- self-hosted runner как код и workflow описан, но не был полноценно прогнан end-to-end
- pipeline завязан на локальную машину
- values file меняется в текущей ветке, что удобно для лабки, но не всегда идеально для production branching strategy

## Итог По Блоку

Блок 5 закрыт как локальная CI/CD схема:

- сервисы пакуются в `Helm`
- образы собираются через `Kaniko`
- образы пушатся в локальный registry
- image tag обновляется через Git
- `Argo CD` забирает изменение и доводит кластер до нового состояния

Это хороший и логичный ответ на ТЗ, потому что он показывает полную цепочку delivery, а не только “мы написали workflow файл”.
