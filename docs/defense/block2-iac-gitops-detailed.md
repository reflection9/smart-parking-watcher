# Блок 2. IaC И GitOps

## Что Требовалось По ТЗ

Во втором блоке требовалось:

- описать базовую инфраструктуру кластера через IaC
- развернуть GitOps-контур
- описать автоматизированный деплой Kafka через Ansible Role

## Что Реализовано

Блок разбит на три слоя:

- `Terraform bootstrap`
- `Argo CD App of Apps`
- `Ansible Role` для `Strimzi/Kafka`

Ключевые пути:

- [infra/terraform/bootstrap](C:/Users/Mikolgi/GolandProjects/smart-parking-watcher/infra/terraform/bootstrap)
- [infra/gitops/apps](C:/Users/Mikolgi/GolandProjects/smart-parking-watcher/infra/gitops/apps)
- [scripts/k8s/bootstrap-argocd.ps1](C:/Users/Mikolgi/GolandProjects/smart-parking-watcher/scripts/k8s/bootstrap-argocd.ps1)
- [ansible/roles/strimzi_kafka](C:/Users/Mikolgi/GolandProjects/smart-parking-watcher/ansible/roles/strimzi_kafka)
- [ansible/playbooks/deploy-strimzi-kafka.yml](C:/Users/Mikolgi/GolandProjects/smart-parking-watcher/ansible/playbooks/deploy-strimzi-kafka.yml)

## Часть 2.1. Terraform Bootstrap

### Что Именно Делает Terraform

`Terraform` не пытается управлять всем приложением. Он делает именно базовый bootstrap кластера:

- namespace `smart-parking`
- namespace `argocd`
- namespace `kafka`
- namespace `observability`
- service account для runtime
- service account для migrations
- service account для bootstrap-а `Argo CD`
- базовый secret `smart-parking-app-secret`

Это видно в:

- [infra/terraform/bootstrap/main.tf](C:/Users/Mikolgi/GolandProjects/smart-parking-watcher/infra/terraform/bootstrap/main.tf)
- [infra/terraform/bootstrap/variables.tf](C:/Users/Mikolgi/GolandProjects/smart-parking-watcher/infra/terraform/bootstrap/variables.tf)

### Почему Terraform Ограничен Bootstrap-Ролью

Это важное архитектурное решение.

Мы не тянем в `Terraform` все Kubernetes-ресурсы приложения, потому что:

- прикладные манифесты удобнее держать в GitOps-контуре
- `Argo CD` лучше подходит для постоянной reconciliation-логики
- смешивать “bootstrap infrastructure” и “day-2 application delivery” в одном слое хуже с точки зрения поддержки

То есть `Terraform` у нас отвечает не за все подряд, а именно за начальный platform baseline.

## Часть 2.2. Argo CD И App Of Apps

### Как Устроен GitOps Слой

`Argo CD` ставится bootstrap-скриптом, который:

- создает namespace `argocd`
- ставит сам `Argo CD`
- рендерит root application с текущим Git remote и текущей веткой
- применяет root app в кластер

После этого начинает работать схема `App of Apps`.

### Root Application

Root app:

- `smart-parking-root`

Он смотрит в:

- [infra/gitops/apps](C:/Users/Mikolgi/GolandProjects/smart-parking-watcher/infra/gitops/apps)

### Child Applications

Child apps:

- `smart-parking-infra`
- `smart-parking-migrations`
- `smart-parking-services`
- позже сюда же добавлялись platform-приложения для `Istio`, `Ingress`, `Jaeger`, observability

### Как Это Работает В Рантайме

Логика такая:

1. Разработчик пушит изменения в Git
2. `Argo CD` видит новую ревизию
3. Root app перечитывает child apps
4. Child apps синхронизируют свои участки кластера
5. Если объект в кластере “уплыл” от desired state, `selfHeal` возвращает его обратно

Именно это и есть GitOps:

- Git является источником истины
- кластер не считается главным источником состояния

### Почему `App of Apps`

Потому что у нас есть несколько логических доменов:

- infra
- migrations
- services
- platform
- observability

Разделение на child apps дает:

- понятные sync wave
- более чистую диагностику
- возможность отдельно смотреть статус разных подсистем

## Часть 2.3. Ansible Role Для Kafka

### Что Делает Role

Role:

- убеждается, что namespace `kafka` существует
- ставит `Strimzi Cluster Operator`
- ждет, пока operator станет `Available`
- рендерит Kafka custom resources из шаблона
- применяет Kafka CR
- ждет, пока `Kafka` CR станет `Ready=True`

Это реализовано в:

- [ansible/roles/strimzi_kafka/tasks/main.yml](C:/Users/Mikolgi/GolandProjects/smart-parking-watcher/ansible/roles/strimzi_kafka/tasks/main.yml)
- [ansible/roles/strimzi_kafka/defaults/main.yml](C:/Users/Mikolgi/GolandProjects/smart-parking-watcher/ansible/roles/strimzi_kafka/defaults/main.yml)
- [ansible/roles/strimzi_kafka/templates/kafka-cluster.yaml.j2](C:/Users/Mikolgi/GolandProjects/smart-parking-watcher/ansible/roles/strimzi_kafka/templates/kafka-cluster.yaml.j2)

### Почему `Strimzi`

Потому что `Strimzi`:

- это стандартный операторный способ жить с Kafka в Kubernetes
- нормально управляет Kafka CRD-моделью
- удобен для объяснения на защите
- лучше, чем вручную городить Kafka StatefulSet как “обычное приложение”

### Какой Kafka Режим Использован

Важно: тут не старый ZooKeeper-вариант.

Используется:

- `KafkaNodePool`
- `KRaft`
- внутренний listener
- single-node / single-replica friendly параметры для локального стенда

То есть архитектура уже ближе к современной Kafka, а не к legacy deployment pattern.

## Почему Вообще Смешаны Terraform, Argo CD И Ansible

Это один из самых частых вопросов.

Ответ:

- `Terraform` отвечает за кластерный bootstrap
- `Argo CD` отвечает за GitOps reconciliation
- `Ansible` отвечает за процедурную автоматизацию operator-based деплоя Kafka

Это не дублирование, а разделение ответственности.

## Коварные Вопросы И Ответы

### Почему не сделать все через Helm

Потому что `Helm` это packaging/template tool, а не full GitOps control plane и не bootstrap IaC engine.

### Почему не сделать все через Argo CD

Потому что `Argo CD` хорош для reconciliation манифестов, но не заменяет удобный bootstrap слой для namespaces, service accounts и sensitive base config.

### Почему не сделать Kafka просто YAML-ами

Потому что Kafka в Kubernetes лучше объяснять и поддерживать через операторную модель, а `Strimzi` как раз это и дает.

### Terraform хранит секреты в state, это нормально

Это допустимо для локальной учебной среды, но не лучший production pattern. Для production нужен более аккуратный подход к secret management.

## Честные Ограничения

- `Terraform` state содержит значения секретов
- `Ansible` слой написан и логически собран, но не позиционируется как полностью production-grade delivery pipeline
- migrations в виде `Job` требуют аккуратного повторного запуска

## Итог По Блоку

Блок 2 закрыт как сочетание:

- IaC bootstrap
- GitOps synchronization
- operator-based Kafka automation

Это хороший архитектурный ответ на ТЗ, потому что каждый инструмент используется в своей сильной роли.
