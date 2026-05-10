const state = {
  userId: null,
  zoneId: null,
  spotId: null,
  reservationId: null,
  reservationStatus: null,
  zones: [],
  notifications: [],
};

const $ = (selector, scope = document) => scope.querySelector(selector);
const $$ = (selector, scope = document) => Array.from(scope.querySelectorAll(selector));

const elements = {
  payload: $("#payload-viewer"),
  eventLog: $("#event-log"),
  toastStack: $("#toast-stack"),
  zonesGrid: $("#zones-grid"),
  notificationsPreview: $("#notifications-preview"),
  selectionSummary: $("#selection-summary"),
  zonesCount: $("#zones-count"),
  notificationsCount: $("#notifications-count"),
  availableSpotsCount: $("#available-spots-count"),
  currentReservationStatus: $("#current-reservation-status"),
  lastSync: $("#last-sync-label"),
  atlasSection: $("#atlas-section"),
  expandZonesButton: $("#expand-zones-button"),
};

function nowLabel() {
  return new Date().toLocaleTimeString("ru-RU");
}

function escapeHtml(value) {
  return String(value)
    .replaceAll("&", "&amp;")
    .replaceAll("<", "&lt;")
    .replaceAll(">", "&gt;")
    .replaceAll('"', "&quot;")
    .replaceAll("'", "&#39;");
}

function stringifyPayload(payload) {
  return typeof payload === "string" ? payload : JSON.stringify(payload, null, 2);
}

function setPayload(payload) {
  elements.payload.textContent = stringifyPayload(payload);
}

function showToast(kind, title, message) {
  const toast = document.createElement("article");
  toast.className = `toast ${kind}`;
  toast.innerHTML = `<strong>${escapeHtml(title)}</strong><p>${escapeHtml(message)}</p>`;
  elements.toastStack.appendChild(toast);
  window.setTimeout(() => toast.remove(), 3600);
}

function pushLog(kind, title, payload) {
  const entry = document.createElement("article");
  entry.className = `log-entry ${kind}`;
  entry.innerHTML = `
    <div class="log-title">
      <span>${escapeHtml(title)}</span>
      <span>${escapeHtml(nowLabel())}</span>
    </div>
    <pre>${escapeHtml(stringifyPayload(payload))}</pre>
  `;

  if ($(".preview-empty", elements.eventLog)) {
    elements.eventLog.innerHTML = "";
  }

  elements.eventLog.prepend(entry);
}

async function api(path, options = {}) {
  const config = {
    method: options.method || "GET",
    headers: {
      ...(options.body !== undefined ? { "Content-Type": "application/json" } : {}),
      ...(options.headers || {}),
    },
  };

  if (options.body !== undefined) {
    config.body = typeof options.body === "string" ? options.body : JSON.stringify(options.body);
  }

  const response = await fetch(path, config);
  const text = await response.text();
  let payload = {};

  if (text) {
    try {
      payload = JSON.parse(text);
    } catch {
      payload = text;
    }
  }

  if (!response.ok) {
    const message =
      typeof payload === "string"
        ? payload
        : payload?.error || payload?.message || `HTTP ${response.status}`;
    throw new Error(message);
  }

  return payload;
}

function parseNumber(value) {
  const raw = String(value ?? "").trim();
  if (!raw) return null;
  const parsed = Number(raw);
  return Number.isFinite(parsed) ? parsed : null;
}

function parseIDs(value) {
  return String(value ?? "")
    .split(",")
    .map((item) => parseNumber(item))
    .filter((item) => item !== null);
}

function buildQuery(params) {
  const query = new URLSearchParams();

  Object.entries(params).forEach(([key, value]) => {
    if (value === null || value === undefined || value === "") return;
    query.set(key, String(value));
  });

  const output = query.toString();
  return output ? `?${output}` : "";
}

function normalizeStatus(status) {
  return String(status || "unknown").toLowerCase();
}

function statusPill(status) {
  return `<span class="status-pill ${normalizeStatus(status)}">${escapeHtml(status || "unknown")}</span>`;
}

function currentZone() {
  return state.zones.find((zone) => Number(zone.id) === Number(state.zoneId)) || null;
}

function currentSpot() {
  const zone = currentZone();
  if (!zone || !Array.isArray(zone.spots)) return null;
  return zone.spots.find((spot) => Number(spot.id) === Number(state.spotId)) || null;
}

function setContext(partial = {}) {
  if (partial.userId !== undefined) state.userId = partial.userId;
  if (partial.zoneId !== undefined) state.zoneId = partial.zoneId;
  if (partial.spotId !== undefined) state.spotId = partial.spotId;
  if (partial.reservationId !== undefined) state.reservationId = partial.reservationId;
  if (partial.reservationStatus !== undefined) state.reservationStatus = partial.reservationStatus;

  renderState();
  prefillInputs();
}

function prefillInputs(force = false) {
  const snapshot = {
    userId: state.userId,
    zoneId: state.zoneId,
    spotId: state.spotId,
    reservationId: state.reservationId,
  };

  $$("[data-prefill]").forEach((input) => {
    const key = input.dataset.prefill;
    const value = snapshot[key];
    if (value === null || value === undefined || value === "") return;
    if (force || !String(input.value || "").trim()) {
      input.value = String(value);
    }
  });
}

function renderSelectionSummary() {
  const zone = currentZone();
  const spot = currentSpot();

  if (!zone) {
    elements.selectionSummary.innerHTML =
      "<p>Выбери зону или место на карте, и мы автоматически подставим id в формы ниже.</p>";
    return;
  }

  elements.selectionSummary.innerHTML = `
    <strong>${escapeHtml(zone.name || `Зона #${zone.id}`)}</strong>
    <p>Zone ID: ${zone.id}. Всего мест: ${zone.total_spots ?? 0}. Свободно: ${zone.available_spots ?? 0}.</p>
    ${
      spot
        ? `<p>Выбрано место <strong>${escapeHtml(spot.number || `#${spot.id}`)}</strong> со статусом ${statusPill(spot.status)}</p>`
        : "<p>Нажми на ячейку, чтобы выбрать конкретное место для брони, истории или ручного изменения статуса.</p>"
    }
  `;
}

function zoneCardMarkup(zone) {
  const spots = Array.isArray(zone.spots) ? [...zone.spots] : [];
  spots.sort((left, right) => String(left.number).localeCompare(String(right.number), "ru"));

  const spotMarkup =
    spots.length > 0
      ? spots
          .map((spot) => {
            const status = normalizeStatus(spot.status);
            const selected = Number(spot.id) === Number(state.spotId) ? "selected" : "";
            return `
              <button
                type="button"
                class="spot-cell ${status} ${selected}"
                data-zone-id="${zone.id}"
                data-spot-id="${spot.id}"
              >
                <div>
                  <div class="spot-number">${escapeHtml(spot.number || `#${spot.id}`)}</div>
                  <div class="spot-status">${escapeHtml(spot.status)}</div>
                </div>
                <div>
                  ${status === "free" ? '<div class="parking-lines" aria-hidden="true"></div>' : '<span class="car-icon" aria-hidden="true"></span>'}
                </div>
              </button>
            `;
          })
          .join("")
      : '<div class="preview-empty">В этой зоне пока нет парковочных мест.</div>';

  return `
    <article class="zone-card" data-zone-id="${zone.id}">
      <div class="zone-head">
        <div>
          <div class="zone-chip">zone #${zone.id}</div>
          <h3>${escapeHtml(zone.name || `Зона #${zone.id}`)}</h3>
          <div class="zone-meta">
            <span>всего: ${zone.total_spots ?? 0}</span>
            <span>свободно: ${zone.available_spots ?? 0}</span>
          </div>
        </div>
        ${statusPill(zone.available_spots > 0 ? "FREE" : "OCCUPIED")}
      </div>
      <div class="spots-grid">${spotMarkup}</div>
      <div class="zone-actions">
        <button class="button button-secondary" type="button" data-zone-action="select" data-zone-id="${zone.id}">
          Выбрать зону
        </button>
        <button class="button button-soft" type="button" data-zone-action="show" data-zone-id="${zone.id}">
          Показать JSON
        </button>
      </div>
    </article>
  `;
}

function renderZones() {
  if (!state.zones.length) {
    elements.zonesGrid.innerHTML = `
      <article class="empty-card">
        <h3>Карта пока пустая</h3>
        <p>Создай первую зону или нажми «Загрузить зоны», и мы покажем сетку мест с машинами на занятых ячейках.</p>
      </article>
    `;
    return;
  }

  elements.zonesGrid.innerHTML = state.zones.map(zoneCardMarkup).join("");
}

function renderNotificationsPreview() {
  if (!state.notifications.length) {
    elements.notificationsPreview.innerHTML =
      '<div class="preview-empty">Уведомления появятся здесь после подписок, освобождения мест или ручного dispatch.</div>';
    return;
  }

  elements.notificationsPreview.innerHTML = state.notifications
    .slice(0, 6)
    .map(
      (notification, index) => `
        <article class="preview-item" data-notification-index="${index}">
          <strong>${escapeHtml(notification.recipient_email || "unknown@email")}</strong>
          <p>${escapeHtml(notification.subject || "Без темы")}</p>
          <small>${statusPill(notification.status)} zone ${notification.zone_id ?? "?"}</small>
        </article>
      `,
    )
    .join("");
}

function renderState() {
  $$("[data-display]").forEach((node) => {
    const key = node.dataset.display;
    const value =
      key === "reservationStatus"
        ? state.reservationStatus || "-"
        : state[key] ?? "-";
    node.textContent = value === null || value === undefined || value === "" ? "-" : String(value);
  });

  const availableSpots = state.zones.reduce(
    (sum, zone) => sum + Number(zone.available_spots || 0),
    0,
  );

  elements.zonesCount.textContent = String(state.zones.length);
  elements.notificationsCount.textContent = String(state.notifications.length);
  elements.availableSpotsCount.textContent = String(availableSpots);
  elements.currentReservationStatus.textContent = state.reservationStatus || "нет";

  renderZones();
  renderNotificationsPreview();
  renderSelectionSummary();
}

function setBusy(button, busy, label) {
  if (!button) return;
  if (!button.dataset.defaultLabel) {
    button.dataset.defaultLabel = button.textContent.trim();
  }
  button.disabled = busy;
  button.textContent = busy ? label : button.dataset.defaultLabel;
}

function resetFormFields(form) {
  form.reset();
  prefillInputs();
}

async function performAction(config, executor) {
  const {
    button = null,
    pendingLabel = "Выполняю...",
    title = "Готово",
    successMessage = "Операция выполнена",
    resetForm = null,
    silent = false,
    afterSuccess = null,
  } = config;

  try {
    setBusy(button, true, pendingLabel);
    const payload = await executor();

    if (!silent && payload !== undefined) {
      setPayload(payload);
      pushLog("success", title, payload);
      showToast(
        "success",
        title,
        typeof successMessage === "function" ? successMessage(payload) : successMessage,
      );
    }

    if (resetForm) {
      resetFormFields(resetForm);
    }

    if (afterSuccess) {
      await afterSuccess(payload);
    }

    return payload;
  } catch (error) {
    const message = error instanceof Error ? error.message : "Неизвестная ошибка";
    setPayload({ error: message });
    pushLog("error", title, message);
    showToast("error", title, message);
    return null;
  } finally {
    setBusy(button, false, pendingLabel);
  }
}

async function fetchZoneAtlas({ button = null, silent = false } = {}) {
  return performAction(
    {
      button,
      pendingLabel: "Гружу зоны...",
      title: "Зоны загружены",
      successMessage: (payload) => `Зон в атласе: ${Array.isArray(payload) ? payload.length : 0}`,
      silent,
      afterSuccess: (payload) => {
        state.zones = Array.isArray(payload) ? payload : [];
        elements.lastSync.textContent = `карта обновлена в ${nowLabel()}`;
        renderState();
      },
    },
    async () => {
      const zones = await api("/zones");
      const detailed = await Promise.all(
        (Array.isArray(zones) ? zones : []).map((zone) =>
          api(`/zones/${zone.id}`).catch(() => zone),
        ),
      );
      return detailed;
    },
  );
}

async function fetchNotifications(filters = {}, { button = null, silent = false } = {}) {
  const query = buildQuery({
    user_id: filters.userId,
    zone_id: filters.zoneId,
    event_id: filters.eventId,
    status: filters.status,
    limit: filters.limit,
  });

  return performAction(
    {
      button,
      pendingLabel: "Загружаю...",
      title: "Уведомления загружены",
      successMessage: (payload) => `Найдено уведомлений: ${Array.isArray(payload) ? payload.length : 0}`,
      silent,
      afterSuccess: (payload) => {
        state.notifications = Array.isArray(payload) ? payload : [];
        renderState();
      },
    },
    () => api(`/notifications${query}`),
  );
}

async function refreshDashboard({ silent = false } = {}) {
  await Promise.all([
    fetchZoneAtlas({ silent }),
    fetchNotifications({ limit: 20 }, { silent }),
  ]);
}

async function pollReservation(reservationId, { attempts = 8, interval = 700 } = {}) {
  let snapshot = null;

  for (let index = 0; index < attempts; index += 1) {
    snapshot = await api(`/reservations/${reservationId}`);
    setContext({
      reservationId: snapshot.id,
      reservationStatus: snapshot.status,
      userId: snapshot.user_id,
      zoneId: snapshot.zone_id,
      spotId: snapshot.spot_id,
    });

    if (!["PENDING", "CONFIRMING", "CANCELLING", "EXPIRING"].includes(snapshot.status)) {
      return snapshot;
    }

    await new Promise((resolve) => window.setTimeout(resolve, interval));
  }

  return snapshot;
}

function readFormNumbers(form, fields) {
  return Object.fromEntries(
    fields.map((field) => [field, parseNumber(form.elements[field]?.value)]),
  );
}

function wireForms() {
  $("#user-register-form").addEventListener("submit", async (event) => {
    event.preventDefault();
    const form = event.currentTarget;
    const button = $('button[type="submit"]', form);
    const body = {
      name: form.elements.name.value.trim(),
      email: form.elements.email.value.trim(),
      password: form.elements.password.value,
    };

    await performAction(
      {
        button,
        title: "Пользователь создан",
        successMessage: "Новый пользователь сохранен в user-service",
        resetForm: form,
        afterSuccess: (payload) => setContext({ userId: payload.id }),
      },
      () => api("/users/register", { method: "POST", body }),
    );
  });

  $("#user-login-form").addEventListener("submit", async (event) => {
    event.preventDefault();
    const form = event.currentTarget;
    const button = $('button[type="submit"]', form);

    await performAction(
      {
        button,
        title: "Вход выполнен",
        successMessage: "Пользователь успешно найден и авторизован",
        afterSuccess: (payload) => setContext({ userId: payload.id }),
      },
      () =>
        api("/users/login", {
          method: "POST",
          body: {
            email: form.elements.email.value.trim(),
            password: form.elements.password.value,
          },
        }),
    );
  });

  $("#user-get-form").addEventListener("submit", async (event) => {
    event.preventDefault();
    const form = event.currentTarget;
    const button = $('button[type="submit"]', form);
    const userId = parseNumber(form.elements.user_id.value);

    await performAction(
      {
        button,
        title: "Пользователь загружен",
        successMessage: `Пользователь ${userId} найден`,
        afterSuccess: (payload) => setContext({ userId: payload.id }),
      },
      () => api(`/users/${userId}`),
    );
  });

  $("#zone-create-form").addEventListener("submit", async (event) => {
    event.preventDefault();
    const form = event.currentTarget;
    const button = $('button[type="submit"]', form);

    await performAction(
      {
        button,
        title: "Зона создана",
        successMessage: "Новая зона добавлена в parking-service",
        resetForm: form,
        afterSuccess: async (payload) => {
          setContext({ zoneId: payload.id, spotId: null });
          await fetchZoneAtlas({ silent: true });
        },
      },
      () =>
        api("/zones", {
          method: "POST",
          body: { name: form.elements.name.value.trim() },
        }),
    );
  });

  $("#zone-get-form").addEventListener("submit", async (event) => {
    event.preventDefault();
    const form = event.currentTarget;
    const button = $('button[type="submit"]', form);
    const zoneId = parseNumber(form.elements.zone_id.value);

    await performAction(
      {
        button,
        title: "Зона загружена",
        successMessage: `Зона ${zoneId} открыта`,
        afterSuccess: (payload) => setContext({ zoneId: payload.id }),
      },
      () => api(`/zones/${zoneId}`),
    );
  });

  $("#zone-list-button").addEventListener("click", () =>
    fetchZoneAtlas({ button: $("#zone-list-button") }),
  );

  $("#spot-create-form").addEventListener("submit", async (event) => {
    event.preventDefault();
    const form = event.currentTarget;
    const button = $('button[type="submit"]', form);
    const zoneId = parseNumber(form.elements.zone_id.value);
    const desiredStatus = form.elements.status.value;

    await performAction(
      {
        button,
        title: "Место добавлено",
        successMessage: "Новое парковочное место готово",
        resetForm: form,
        afterSuccess: async (payload) => {
          setContext({ zoneId: payload.zone_id, spotId: payload.id });
          await fetchZoneAtlas({ silent: true });
        },
      },
      async () => {
        const spot = await api(`/zones/${zoneId}/spots`, {
          method: "POST",
          body: { number: form.elements.number.value.trim() },
        });

        if (desiredStatus && desiredStatus !== "FREE") {
          await api(`/zones/${zoneId}/spots/${spot.id}/status`, {
            method: "PATCH",
            body: { status: desiredStatus },
          });
          return api(`/spots/${spot.id}/zones/${zoneId}`);
        }

        return spot;
      },
    );
  });

  $("#spot-get-form").addEventListener("submit", async (event) => {
    event.preventDefault();
    const form = event.currentTarget;
    const button = $('button[type="submit"]', form);
    const { zone_id: zoneId, spot_id: spotId } = readFormNumbers(form, ["zone_id", "spot_id"]);

    await performAction(
      {
        button,
        title: "Место загружено",
        successMessage: `Место ${spotId} открыто`,
        afterSuccess: (payload) => setContext({ zoneId: payload.zone_id, spotId: payload.id }),
      },
      () => api(`/spots/${spotId}/zones/${zoneId}`),
    );
  });

  $("#spot-status-form").addEventListener("submit", async (event) => {
    event.preventDefault();
    const form = event.currentTarget;
    const button = $('button[type="submit"]', form);
    const { zone_id: zoneId, spot_id: spotId } = readFormNumbers(form, ["zone_id", "spot_id"]);

    await performAction(
      {
        button,
        title: "Статус места обновлен",
        successMessage: "Ручной статус успешно изменен",
        afterSuccess: async (payload) => {
          setContext({ zoneId: payload.zone_id, spotId: payload.id });
          await Promise.all([
            fetchZoneAtlas({ silent: true }),
            fetchNotifications({ limit: 20 }, { silent: true }),
          ]);
        },
      },
      () =>
        api(`/zones/${zoneId}/spots/${spotId}/status`, {
          method: "PATCH",
          body: { status: form.elements.status.value },
        }),
    );
  });

  $("#spot-transition-form").addEventListener("click", async (event) => {
    const button = event.target.closest("[data-transition]");
    if (!button) return;

    const form = $("#spot-transition-form");
    const { zone_id: zoneId, spot_id: spotId } = readFormNumbers(form, ["zone_id", "spot_id"]);
    const transition = button.dataset.transition;

    await performAction(
      {
        button,
        title: `Переход ${transition}`,
        successMessage: `Операция ${transition} выполнена`,
        afterSuccess: async (payload) => {
          setContext({ zoneId: payload.zone_id, spotId: payload.id });
          await Promise.all([
            fetchZoneAtlas({ silent: true }),
            fetchNotifications({ limit: 20 }, { silent: true }),
          ]);
        },
      },
      () => api(`/zones/${zoneId}/spots/${spotId}/${transition}`, { method: "POST" }),
    );
  });

  $("#subscription-create-form").addEventListener("submit", async (event) => {
    event.preventDefault();
    const form = event.currentTarget;
    const button = $('button[type="submit"]', form);
    const { user_id: userId, zone_id: zoneId } = readFormNumbers(form, ["user_id", "zone_id"]);

    await performAction(
      {
        button,
        title: "Подписка создана",
        successMessage: "Пользователь подписан на зону",
        resetForm: form,
        afterSuccess: async () => {
          setContext({ userId, zoneId });
          await fetchNotifications({ userId, limit: 20 }, { silent: true });
        },
      },
      () =>
        api("/subscriptions", {
          method: "POST",
          body: { user_id: userId, zone_id: zoneId },
        }),
    );
  });

  $("#subscription-list-user-form").addEventListener("submit", async (event) => {
    event.preventDefault();
    const form = event.currentTarget;
    const button = $('button[type="submit"]', form);
    const userId = parseNumber(form.elements.user_id.value);

    await performAction(
      {
        button,
        title: "Подписки пользователя",
        successMessage: `Получены подписки user ${userId}`,
        afterSuccess: () => setContext({ userId }),
      },
      () => api(`/subscriptions/users/${userId}`),
    );
  });

  $("#subscription-list-zone-form").addEventListener("submit", async (event) => {
    event.preventDefault();
    const form = event.currentTarget;
    const button = $('button[type="submit"]', form);
    const zoneId = parseNumber(form.elements.zone_id.value);

    await performAction(
      {
        button,
        title: "Подписчики зоны",
        successMessage: `Получен список подписчиков зоны ${zoneId}`,
        afterSuccess: () => setContext({ zoneId }),
      },
      () => api(`/subscriptions/zones/${zoneId}`),
    );
  });

  $("#subscription-delete-form").addEventListener("submit", async (event) => {
    event.preventDefault();
    const form = event.currentTarget;
    const button = $('button[type="submit"]', form);
    const { user_id: userId, zone_id: zoneId } = readFormNumbers(form, ["user_id", "zone_id"]);

    await performAction(
      {
        button,
        title: "Подписка удалена",
        successMessage: "Пользователь успешно отписан от зоны",
        afterSuccess: async () => {
          setContext({ userId, zoneId });
          await fetchNotifications({ userId, limit: 20 }, { silent: true });
        },
      },
      () => api(`/subscriptions/users/${userId}/zones/${zoneId}`, { method: "DELETE" }),
    );
  });

  $("#reservation-create-form").addEventListener("submit", async (event) => {
    event.preventDefault();
    const form = event.currentTarget;
    const button = $('button[type="submit"]', form);
    const { user_id: userId, zone_id: zoneId, spot_id: spotId } = readFormNumbers(form, [
      "user_id",
      "zone_id",
      "spot_id",
    ]);

    await performAction(
      {
        button,
        pendingLabel: "Запускаю сагу...",
        title: "Бронь создана",
        successMessage: "Сага бронирования стартовала",
        resetForm: form,
        afterSuccess: async (payload) => {
          setContext({
            userId,
            zoneId,
            spotId,
            reservationId: payload.id,
            reservationStatus: payload.status,
          });

          const snapshot = await pollReservation(payload.id);
          if (snapshot) setPayload(snapshot);

          await Promise.all([
            fetchZoneAtlas({ silent: true }),
            fetchNotifications({ userId, limit: 20 }, { silent: true }),
          ]);
        },
      },
      () =>
        api("/reservations", {
          method: "POST",
          body: { user_id: userId, zone_id: zoneId, spot_id: spotId },
        }),
    );
  });

  $("#reservation-get-form").addEventListener("submit", async (event) => {
    event.preventDefault();
    const form = event.currentTarget;
    const button = $('button[type="submit"]', form);
    const reservationId = parseNumber(form.elements.reservation_id.value);

    await performAction(
      {
        button,
        title: "Бронь загружена",
        successMessage: `Бронь ${reservationId} получена`,
        afterSuccess: (payload) =>
          setContext({
            reservationId: payload.id,
            reservationStatus: payload.status,
            userId: payload.user_id,
            zoneId: payload.zone_id,
            spotId: payload.spot_id,
          }),
      },
      () => api(`/reservations/${reservationId}`),
    );
  });

  $("#reservation-list-user-form").addEventListener("submit", async (event) => {
    event.preventDefault();
    const form = event.currentTarget;
    const button = $('button[type="submit"]', form);
    const userId = parseNumber(form.elements.user_id.value);

    await performAction(
      {
        button,
        title: "Брони пользователя",
        successMessage: `Список броней пользователя ${userId} получен`,
        afterSuccess: () => setContext({ userId }),
      },
      () => api(`/reservations/users/${userId}`),
    );
  });

  $("#reservation-action-form").addEventListener("click", async (event) => {
    const button = event.target.closest("[data-reservation-action]");
    if (!button) return;

    const form = $("#reservation-action-form");
    const reservationId = parseNumber(form.elements.reservation_id.value);
    const action = button.dataset.reservationAction;

    await performAction(
      {
        button,
        pendingLabel: "Провожу переход...",
        title: `Бронь: ${action}`,
        successMessage: `Переход ${action} отработал`,
        afterSuccess: async (payload) => {
          setContext({
            reservationId: payload.id,
            reservationStatus: payload.status,
            userId: payload.user_id,
            zoneId: payload.zone_id,
            spotId: payload.spot_id,
          });

          const snapshot = await pollReservation(payload.id);
          if (snapshot) setPayload(snapshot);

          await Promise.all([
            fetchZoneAtlas({ silent: true }),
            fetchNotifications({ userId: payload.user_id, limit: 20 }, { silent: true }),
          ]);
        },
      },
      () => api(`/reservations/${reservationId}/${action}`, { method: "POST" }),
    );
  });

  $("#history-create-form").addEventListener("submit", async (event) => {
    event.preventDefault();
    const form = event.currentTarget;
    const button = $('button[type="submit"]', form);
    const { zone_id: zoneId, spot_id: spotId } = readFormNumbers(form, ["zone_id", "spot_id"]);

    await performAction(
      {
        button,
        title: "Историческое событие создано",
        successMessage: "Событие записано в warm-хранилище",
        resetForm: form,
        afterSuccess: () => setContext({ zoneId, spotId }),
      },
      () =>
        api("/history/events", {
          method: "POST",
          body: {
            zone_id: zoneId,
            spot_id: spotId,
            event_type: form.elements.event_type.value.trim(),
            old_status: form.elements.old_status.value.trim(),
            new_status: form.elements.new_status.value.trim(),
            source: form.elements.source.value.trim(),
          },
        }),
    );
  });

  $("#history-zone-form").addEventListener("submit", async (event) => {
    event.preventDefault();
    const form = event.currentTarget;
    const button = $('button[type="submit"]', form);
    const zoneId = parseNumber(form.elements.zone_id.value);

    await performAction(
      {
        button,
        title: "История зоны",
        successMessage: `История зоны ${zoneId} получена`,
        afterSuccess: () => setContext({ zoneId }),
      },
      () => api(`/history/zones/${zoneId}`),
    );
  });

  $("#history-spot-form").addEventListener("submit", async (event) => {
    event.preventDefault();
    const form = event.currentTarget;
    const button = $('button[type="submit"]', form);
    const spotId = parseNumber(form.elements.spot_id.value);

    await performAction(
      {
        button,
        title: "История места",
        successMessage: `История места ${spotId} получена`,
        afterSuccess: () => setContext({ spotId }),
      },
      () => api(`/history/spots/${spotId}`),
    );
  });

  $("#history-reservation-form").addEventListener("submit", async (event) => {
    event.preventDefault();
    const form = event.currentTarget;
    const button = $('button[type="submit"]', form);
    const reservationId = parseNumber(form.elements.reservation_id.value);

    await performAction(
      {
        button,
        title: "История брони",
        successMessage: `История брони ${reservationId} получена`,
        afterSuccess: () => setContext({ reservationId }),
      },
      () => api(`/history/reservations/${reservationId}`),
    );
  });

  $("#history-archive-form").addEventListener("submit", async (event) => {
    event.preventDefault();
    const form = event.currentTarget;
    const button = $('button[type="submit"]', form);
    const olderThanHours = parseNumber(form.elements.older_than_hours.value);

    await performAction(
      {
        button,
        title: "Архивация истории",
        successMessage: "Старые события отправлены в MinIO",
      },
      () => api(`/history/archive${buildQuery({ older_than_hours: olderThanHours })}`, { method: "POST" }),
    );
  });

  $("#notification-dispatch-form").addEventListener("submit", async (event) => {
    event.preventDefault();
    const form = event.currentTarget;
    const button = $('button[type="submit"]', form);
    const zoneId = parseNumber(form.elements.zone_id.value);
    const spotId = parseNumber(form.elements.spot_id.value);
    const userIds = parseIDs(form.elements.user_ids.value);

    await performAction(
      {
        button,
        title: "Ручной dispatch",
        successMessage: "Уведомление вручную отправлено в notification-service",
        afterSuccess: async () => {
          setContext({ zoneId, spotId: spotId ?? state.spotId });
          await fetchNotifications({ limit: 20 }, { silent: true });
        },
      },
      () =>
        api("/notifications/spot-freed", {
          method: "POST",
          body: {
            event_id: form.elements.event_id.value.trim(),
            event_type: form.elements.event_type.value.trim(),
            zone_id: zoneId,
            ...(spotId ? { spot_id: spotId } : {}),
            ...(userIds.length ? { user_ids: userIds } : {}),
          },
        }),
    );
  });

  $("#notification-list-form").addEventListener("submit", async (event) => {
    event.preventDefault();
    const form = event.currentTarget;
    const button = $('button[type="submit"]', form);

    await fetchNotifications(
      {
        userId: parseNumber(form.elements.user_id.value),
        zoneId: parseNumber(form.elements.zone_id.value),
        eventId: form.elements.event_id.value.trim(),
        status: form.elements.status.value,
        limit: parseNumber(form.elements.limit.value),
      },
      { button },
    );
  });

  $("#notification-get-form").addEventListener("submit", async (event) => {
    event.preventDefault();
    const form = event.currentTarget;
    const button = $('button[type="submit"]', form);
    const notificationId = parseNumber(form.elements.notification_id.value);

    await performAction(
      {
        button,
        title: "Уведомление загружено",
        successMessage: `Уведомление ${notificationId} получено`,
      },
      () => api(`/notifications/${notificationId}`),
    );
  });
}

function wireInteractiveUI() {
  $("#refresh-dashboard-button").addEventListener("click", () => refreshDashboard());
  $("#fetch-zones-button").addEventListener("click", () => fetchZoneAtlas({ button: $("#fetch-zones-button") }));
  $("#notifications-fetch-button").addEventListener("click", () =>
    fetchNotifications({ limit: 20 }, { button: $("#notifications-fetch-button") }),
  );
  $("#notifications-preview-refresh").addEventListener("click", () =>
    fetchNotifications({ limit: 20 }, { button: $("#notifications-preview-refresh") }),
  );

  $("#clear-log-button").addEventListener("click", () => {
    elements.eventLog.innerHTML = '<div class="preview-empty">Журнал очищен. Новые действия снова появятся здесь.</div>';
  });

  elements.expandZonesButton.addEventListener("click", () => {
    const expanded = elements.atlasSection.classList.toggle("expanded");
    elements.expandZonesButton.textContent = expanded ? "Свернуть карту" : "Развернуть карту";
  });

  elements.zonesGrid.addEventListener("click", (event) => {
    const spotButton = event.target.closest(".spot-cell");
    if (spotButton) {
      const zoneId = parseNumber(spotButton.dataset.zoneId);
      const spotId = parseNumber(spotButton.dataset.spotId);
      setContext({ zoneId, spotId });
      const spot = currentSpot();
      if (spot) setPayload(spot);
      pushLog("success", "Контекст обновлен по карте", { zone_id: zoneId, spot_id: spotId });
      return;
    }

    const zoneAction = event.target.closest("[data-zone-action]");
    if (!zoneAction) return;

    const zoneId = parseNumber(zoneAction.dataset.zoneId);
    const zone = state.zones.find((item) => Number(item.id) === Number(zoneId));
    if (!zone) return;

    if (zoneAction.dataset.zoneAction === "select") {
      setContext({ zoneId: zone.id, spotId: zone.spots?.[0]?.id ?? null });
      pushLog("success", "Зона выбрана из атласа", zone);
      return;
    }

    setPayload(zone);
  });

  elements.notificationsPreview.addEventListener("click", (event) => {
    const card = event.target.closest("[data-notification-index]");
    if (!card) return;
    const notification = state.notifications[Number(card.dataset.notificationIndex)];
    if (!notification) return;
    setPayload(notification);
  });

  $("#demo-seed-button").addEventListener("click", async (event) => {
    const button = event.currentTarget;
    const stamp = Date.now().toString().slice(-6);

    await performAction(
      {
        button,
        pendingLabel: "Собираю демо...",
        title: "Демо-сценарий готов",
        successMessage: "Созданы пользователь, зона, два места, подписка и бронь",
      },
      async () => {
        const user = await api("/users/register", {
          method: "POST",
          body: {
            name: `Гость ${stamp}`,
            email: `demo-${stamp}@example.com`,
            password: "password123",
          },
        });

        const zone = await api("/zones", {
          method: "POST",
          body: { name: `Зона ${stamp}` },
        });

        const freeSpot = await api(`/zones/${zone.id}/spots`, {
          method: "POST",
          body: { number: `A-${stamp}` },
        });

        const occupiedSpot = await api(`/zones/${zone.id}/spots`, {
          method: "POST",
          body: { number: `B-${stamp}` },
        });

        await api(`/zones/${zone.id}/spots/${occupiedSpot.id}/status`, {
          method: "PATCH",
          body: { status: "OCCUPIED" },
        });

        const subscription = await api("/subscriptions", {
          method: "POST",
          body: { user_id: user.id, zone_id: zone.id },
        });

        const reservation = await api("/reservations", {
          method: "POST",
          body: { user_id: user.id, zone_id: zone.id, spot_id: freeSpot.id },
        });

        setContext({
          userId: user.id,
          zoneId: zone.id,
          spotId: freeSpot.id,
          reservationId: reservation.id,
          reservationStatus: reservation.status,
        });

        const reservationSnapshot = await pollReservation(reservation.id);
        await refreshDashboard({ silent: true });

        return {
          user,
          zone,
          free_spot: freeSpot,
          occupied_spot: occupiedSpot,
          subscription,
          reservation: reservationSnapshot,
        };
      },
    );
  });
}

document.addEventListener("DOMContentLoaded", async () => {
  wireForms();
  wireInteractiveUI();
  renderState();
  prefillInputs();

  try {
    await refreshDashboard({ silent: true });
  } catch (error) {
    pushLog(
      "error",
      "Первичная загрузка",
      error instanceof Error ? error.message : "Не удалось загрузить стартовые данные",
    );
  }
});
