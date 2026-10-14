const STATUSES = {
  waiting: "Ждём",
  accepted: "Оффер",
  rejected: "Отказ",
};

const state = {
  user: null,
  responses: [],
  filter: "all",
  query: "",
};

const notice = document.querySelector("#notice");
const gate = document.querySelector("#gate");
const board = document.querySelector("#board");

function humanError(text) {
  const message = (text || "").trim();
  const known = {
    "user not found": "Такого профиля нет. Создайте новый справа.",
    "username already exists": "Это имя уже занято.",
    "email already exists": "Эта почта уже занята.",
    "username and email are required": "Нужны имя и почта.",
    "user_id, company, and vacancy are required": "Нужны компания и вакансия.",
    "invalid json": "Сервер не принял данные.",
    "internal error": "Сервер не смог выполнить запрос.",
  };
  return known[message] || message || "Не удалось выполнить запрос";
}

function showNotice(message) {
  notice.hidden = !message;
  notice.textContent = message || "";
}

async function api(path, options = {}) {
  const response = await fetch(path, {
    ...options,
    headers: {
      ...(options.body ? { "Content-Type": "application/json" } : {}),
      ...options.headers,
    },
  });
  const text = await response.text();
  if (!response.ok) {
    throw new Error(humanError(text));
  }
  return text ? JSON.parse(text) : null;
}

function saveUser(user) {
  state.user = user;
  localStorage.setItem("responses.user", JSON.stringify(user));
}

function clearUser() {
  state.user = null;
  state.responses = [];
  localStorage.removeItem("responses.user");
  render();
}

function render() {
  const signedIn = Boolean(state.user);
  gate.hidden = signedIn;
  board.hidden = !signedIn;
  if (!signedIn) return;

  document.querySelector("#profile-name").textContent = state.user.Username;
  document.querySelector("#profile-email").textContent = state.user.Email;
  renderStats();
  renderList();
}

function renderStats() {
  const counts = { waiting: 0, accepted: 0, rejected: 0 };
  for (const item of state.responses) counts[item.Status] = (counts[item.Status] || 0) + 1;
  document.querySelector("#stats").innerHTML = ["waiting", "accepted", "rejected"]
    .map((status) => `<div class="stat"><b>${counts[status]}</b><span>${STATUSES[status]}</span></div>`)
    .join("");
}

function visibleResponses() {
  const query = state.query.trim().toLowerCase();
  return state.responses
    .filter((item) => state.filter === "all" || item.Status === state.filter)
    .filter((item) => {
      if (!query) return true;
      return `${item.Company} ${item.Vacancy}`.toLowerCase().includes(query);
    })
    .sort((a, b) => a.Company.localeCompare(b.Company, "ru"));
}

function renderList() {
  const items = visibleResponses();
  const list = document.querySelector("#list");
  if (items.length === 0) {
    list.innerHTML = `<div class="empty">В этом списке пока ничего нет.</div>`;
    return;
  }

  list.innerHTML = items.map((item) => `
    <article class="row" data-id="${item.ID}">
      <div>
        <strong>${escapeHtml(item.Company)}</strong>
        <small>Компания</small>
      </div>
      <div>
        <strong>${escapeHtml(item.Vacancy)}</strong>
        <small>Вакансия</small>
      </div>
      <select class="status ${escapeHtml(item.Status)}" data-action="status">
        ${Object.entries(STATUSES).map(([value, label]) => `
          <option value="${value}" ${value === item.Status ? "selected" : ""}>${label}</option>
        `).join("")}
      </select>
      <button type="button" class="danger" data-action="delete">Удалить</button>
    </article>
  `).join("");
}

function escapeHtml(value) {
  return String(value)
    .replaceAll("&", "&amp;")
    .replaceAll("<", "&lt;")
    .replaceAll(">", "&gt;")
    .replaceAll('"', "&quot;");
}

async function loadResponses() {
  const data = await api(`/api/v1/responses/user/${state.user.ID}`);
  state.responses = Array.isArray(data) ? data : [];
  render();
}

document.querySelector("#login-form").addEventListener("submit", async (event) => {
  event.preventDefault();
  showNotice("");
  const username = new FormData(event.currentTarget).get("username").trim();
  try {
    const user = await api(`/api/v1/users/username/${encodeURIComponent(username)}`);
    saveUser(user);
    await loadResponses();
  } catch (error) {
    showNotice(error.message);
  }
});

document.querySelector("#register-form").addEventListener("submit", async (event) => {
  event.preventDefault();
  showNotice("");
  const formElement = event.currentTarget;
  const form = new FormData(formElement);
  try {
    const user = await api("/api/v1/users", {
      method: "POST",
      body: JSON.stringify({
        username: form.get("username").trim(),
        email: form.get("email").trim(),
      }),
    });
    saveUser(user);
    formElement.reset();
    await loadResponses();
  } catch (error) {
    showNotice(error.message);
  }
});

document.querySelector("#logout").addEventListener("click", clearUser);

document.querySelector("#email-form").addEventListener("submit", async (event) => {
  event.preventDefault();
  showNotice("");
  const formElement = event.currentTarget;
  const email = new FormData(formElement).get("email").trim();
  try {
    await api(`/api/v1/users/${state.user.ID}/email/${encodeURIComponent(email)}`, { method: "PATCH" });
    saveUser({ ...state.user, Email: email });
    formElement.reset();
    render();
  } catch (error) {
    showNotice(error.message);
  }
});

document.querySelector("#response-form").addEventListener("submit", async (event) => {
  event.preventDefault();
  showNotice("");
  const formElement = event.currentTarget;
  const form = new FormData(formElement);
  try {
    await api("/api/v1/responses", {
      method: "POST",
      body: JSON.stringify({
        user_id: state.user.ID,
        company: form.get("company").trim(),
        vacancy: form.get("vacancy").trim(),
        status: form.get("status"),
      }),
    });
    formElement.reset();
    await loadResponses();
  } catch (error) {
    showNotice(error.message);
  }
});

document.querySelector("#search").addEventListener("input", (event) => {
  state.query = event.currentTarget.value;
  renderList();
});

document.querySelector("#filters").addEventListener("click", (event) => {
  const button = event.target.closest("[data-filter]");
  if (!button) return;
  state.filter = button.dataset.filter;
  document.querySelectorAll("#filters .chip").forEach((chip) => {
    chip.classList.toggle("is-on", chip === button);
  });
  renderList();
});

document.querySelector("#list").addEventListener("change", async (event) => {
  const select = event.target.closest("[data-action='status']");
  if (!select) return;
  const row = select.closest("[data-id]");
  showNotice("");
  try {
    await api(`/api/v1/responses/${row.dataset.id}/status/${select.value}`, { method: "PATCH" });
    const item = state.responses.find((response) => response.ID === row.dataset.id);
    if (item) item.Status = select.value;
    render();
  } catch (error) {
    showNotice(error.message);
    renderList();
  }
});

document.querySelector("#list").addEventListener("click", async (event) => {
  const button = event.target.closest("[data-action='delete']");
  if (!button) return;
  const row = button.closest("[data-id]");
  const item = state.responses.find((response) => response.ID === row.dataset.id);
  if (!item || !window.confirm(`Удалить отклик в ${item.Company}?`)) return;
  showNotice("");
  try {
    await api(`/api/v1/responses/${item.ID}`, { method: "DELETE" });
    state.responses = state.responses.filter((response) => response.ID !== item.ID);
    render();
  } catch (error) {
    showNotice(error.message);
  }
});

const saved = localStorage.getItem("responses.user");
if (saved) {
  try {
    state.user = JSON.parse(saved);
    loadResponses().catch((error) => showNotice(error.message));
  } catch {
    localStorage.removeItem("responses.user");
  }
}
