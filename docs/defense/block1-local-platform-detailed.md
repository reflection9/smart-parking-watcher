# Блок 1. Локальная Платформа И Kubernetes

## Что Требовалось По ТЗ

В первом блоке требовалось:

- поднять локальный Kubernetes-кластер
- использовать современный CNI
- обеспечить воспроизводимый локальный runtime для следующих блоков
- рассмотреть автоматическое масштабирование worker-нод

## Что Реализовано

В проекте реализован локальный кластер на основе:

- `k3d`
- `k3s`
- `Cilium`

Основные файлы и скрипты:

- [infra/k8s/local/k3d-cilium-cluster.yaml](C:/Users/Mikolgi/GolandProjects/smart-parking-watcher/infra/k8s/local/k3d-cilium-cluster.yaml)
- [infra/k8s/local/cilium-values.yaml](C:/Users/Mikolgi/GolandProjects/smart-parking-watcher/infra/k8s/local/cilium-values.yaml)
- [scripts/k8s/create-local-cilium-cluster.ps1](C:/Users/Mikolgi/GolandProjects/smart-parking-watcher/scripts/k8s/create-local-cilium-cluster.ps1)
- [scripts/k8s/validate-local-cilium-cluster.ps1](C:/Users/Mikolgi/GolandProjects/smart-parking-watcher/scripts/k8s/validate-local-cilium-cluster.ps1)
- [scripts/k8s/prepare-local-k8s-mode.ps1](C:/Users/Mikolgi/GolandProjects/smart-parking-watcher/scripts/k8s/prepare-local-k8s-mode.ps1)
- [docs/block1-local-platform-runbook.md](C:/Users/Mikolgi/GolandProjects/smart-parking-watcher/docs/block1-local-platform-runbook.md)
- [docs/block1-node-autoscaling-decision.md](C:/Users/Mikolgi/GolandProjects/smart-parking-watcher/docs/block1-node-autoscaling-decision.md)

## Как Это Работает

Локальный runtime устроен так:

- `Docker Desktop` дает контейнерный движок
- `k3d` создает Kubernetes-ноды как Docker-контейнеры
- внутри этих нод работает `k3s`
- `Cilium` поднимает сетевой слой кластера
- `ingress`, `Argo CD`, `Istio`, observability и приложение потом ставятся уже поверх этого кластера

То есть фактически у нас не “виртуальные машины”, а облегченный Kubernetes поверх Docker.

## Почему Выбраны Именно Эти Технологии

### Почему `k3d`

`k3d` хорош для локальной разработки, потому что:

- быстро поднимается и удаляется
- не требует отдельного гипервизора
- хорошо управляется скриптами
- отлично подходит для `GitOps` и platform-лабы

### Почему `k3s`

`k3s` выбран потому что:

- это легкий Kubernetes-дистрибутив
- он быстрее и проще для локальной среды, чем “тяжелый” full-sized control plane
- для учебного и лабораторного стенда он дает почти тот же API surface, что и обычный Kubernetes

### Почему `Cilium`

`Cilium` выбран как CNI, потому что:

- это современный production-grade CNI
- он закрывает требование по сетевым политикам
- он хорошо масштабируется дальше на mesh и observability use cases
- он стандартный и сильный аргумент на защите, когда спрашивают про сетевой слой

## Как Мы Контролируем Локальную Среду

В репозитории есть два режима:

- `docker compose` режим
- `k3d/k8s` режим

Главное правило:

- одновременно они не должны быть активны

Причина:

- `k3d` сам работает поверх Docker
- если одновременно запустить и `docker compose`, и `k3d`, Docker Desktop начинает держать слишком много контейнеров, сетей и volume-ов
- именно это и было причиной, почему локальная машина “задыхалась”

Правильный порядок работы:

1. Включить `Docker Desktop`
2. Убедиться, что `docker compose` стек остановлен
3. Выполнить `prepare-local-k8s-mode.ps1`
4. Поднять кластер через `create-local-cilium-cluster.ps1`
5. Проверить кластер через `validate-local-cilium-cluster.ps1`

## Что С Пунктом 1.2 Про Автомасштабирование Нод

Здесь важно отвечать честно.

Полноценный `Karpenter` или `Cluster Autoscaler` в локальном `k3d`-лабе не ставился.

Причина не в том, что “не успели”, а в том, что:

- `Karpenter` рассчитан на поддерживаемый backend провижининга нод
- `Cluster Autoscaler` работает через cloud/provider-specific node groups
- в `k3d` worker-ноды это просто Docker-контейнеры, а не managed instances

То есть реального облачного node provisioning API здесь нет.

## Чем Закрыт Пункт 1.2

Пункт закрыт как:

- documented limitation
- controlled local scaling

Практически это означает:

- мы можем вручную добавлять и удалять worker-ноды
- мы не называем это “настоящим autoscaling”
- мы явно документируем, почему продовый autoscaler здесь неуместен

Пример:

- `k3d node create ... --role agent`
- `k3d node delete ...`

## Что Говорить На Защите

Если спрашивают “почему не Minikube”, ответ такой:

`Minikube` не нужен, потому что репозиторий уже стандартизирован под `k3d + k3s + Cilium`, и вторая локальная среда только создавала бы путаницу.

Если спрашивают “почему Docker Desktop должен быть включен”, ответ такой:

Потому что `k3d` создает Kubernetes-ноды как Docker-контейнеры, и без Docker сам кластер просто не существует.

Если спрашивают “почему не сделали autoscaling как в облаке”, ответ такой:

В локальном `k3d` нет cloud provider backend-а, поэтому `Karpenter` и `Cluster Autoscaler` здесь не решают реальную задачу. Мы честно зафиксировали ограничение и сделали контролируемое ручное масштабирование worker capacity.

## Честные Ограничения

- это локальный кластер, а не production
- ноды живут как Docker-контейнеры
- autoscaling нод здесь не облачного типа
- при грубых рестартах Docker или `k3d` может понадобиться повторно привести `Cilium` в порядок

## Итог По Блоку

Первый блок закрыт как воспроизводимая локальная Kubernetes-платформа:

- кластер поднимается скриптами
- сетевой слой закрыт `Cilium`
- режим работы документирован
- конфликт `compose` и `k3d` устранен организационно и скриптами
- ограничение по autoscaling не скрыто, а грамотно оформлено
