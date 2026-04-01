const state = {
  apiBase: localStorage.getItem("todo_api_base") || "http://localhost:8080/api/v1",
  token: localStorage.getItem("todo_token") || "",
  selectedListId: null,
  lists: [],
  items: [],
};

const el = {
  apiBase: document.getElementById("apiBase"),
  saveApiBase: document.getElementById("saveApiBase"),
  tabSignIn: document.getElementById("tabSignIn"),
  tabSignUp: document.getElementById("tabSignUp"),
  signInForm: document.getElementById("signInForm"),
  signUpForm: document.getElementById("signUpForm"),
  signInUsername: document.getElementById("signInUsername"),
  signInPassword: document.getElementById("signInPassword"),
  signUpName: document.getElementById("signUpName"),
  signUpUsername: document.getElementById("signUpUsername"),
  signUpPassword: document.getElementById("signUpPassword"),
  logoutBtn: document.getElementById("logoutBtn"),
  tokenPreview: document.getElementById("tokenPreview"),
  refreshLists: document.getElementById("refreshLists"),
  createListForm: document.getElementById("createListForm"),
  newListTitle: document.getElementById("newListTitle"),
  newListDescription: document.getElementById("newListDescription"),
  lists: document.getElementById("lists"),
  itemsTitle: document.getElementById("itemsTitle"),
  refreshItems: document.getElementById("refreshItems"),
  createItemForm: document.getElementById("createItemForm"),
  newItemTitle: document.getElementById("newItemTitle"),
  newItemDescription: document.getElementById("newItemDescription"),
  items: document.getElementById("items"),
  toast: document.getElementById("toast"),
};

function showToast(text) {
  el.toast.textContent = text;
  el.toast.classList.add("show");
  setTimeout(() => el.toast.classList.remove("show"), 1800);
}

function setToken(token) {
  state.token = token || "";
  if (state.token) {
    localStorage.setItem("todo_token", state.token);
  } else {
    localStorage.removeItem("todo_token");
  }
  el.tokenPreview.textContent = `token: ${state.token || "-"}`;
}

function authHeaders() {
  const h = { "Content-Type": "application/json" };
  if (state.token) h.Authorization = `Bearer ${state.token}`;
  return h;
}

async function api(path, options = {}) {
  const url = `${state.apiBase}${path}`;
  const res = await fetch(url, {
    ...options,
    headers: {
      ...(options.headers || {}),
      ...(options.body ? { "Content-Type": "application/json" } : {}),
      ...(state.token ? { Authorization: `Bearer ${state.token}` } : {}),
    },
  });

  const raw = await res.text();
  let body = null;
  if (raw) {
    try { body = JSON.parse(raw); } catch { body = raw; }
  }

  if (!res.ok) {
    const msg = body?.error || body?.message || `HTTP ${res.status}`;
    throw new Error(msg);
  }
  return body;
}

function renderLists() {
  el.lists.innerHTML = "";
  state.lists.forEach((list) => {
    const li = document.createElement("li");
    const active = state.selectedListId === list.id;
    li.innerHTML = `
      <div class="item-title">${list.title}</div>
      <div class="item-meta">${list.description || "без описания"}</div>
      <div class="item-actions">
        <button class="small ${active ? "" : "secondary"}" data-act="open">${active ? "Открыт" : "Открыть"}</button>
        <button class="small secondary" data-act="edit">Изменить</button>
        <button class="small danger" data-act="delete">Удалить</button>
      </div>
    `;

    li.querySelector('[data-act="open"]').onclick = async () => {
      state.selectedListId = list.id;
      el.itemsTitle.textContent = `Задачи: ${list.title}`;
      renderLists();
      await loadItems();
    };

    li.querySelector('[data-act="edit"]').onclick = async () => {
      const title = prompt("Название", list.title);
      if (title === null) return;
      const description = prompt("Описание", list.description || "");
      if (description === null) return;
      await api(`/lists/${list.id}/`, {
        method: "PATCH",
        body: JSON.stringify({ title, description }),
      });
      showToast("Список обновлен");
      await loadLists();
    };

    li.querySelector('[data-act="delete"]').onclick = async () => {
      if (!confirm("Удалить список?")) return;
      await api(`/lists/${list.id}/`, { method: "DELETE" });
      if (state.selectedListId === list.id) {
        state.selectedListId = null;
        state.items = [];
        renderItems();
      }
      showToast("Список удален");
      await loadLists();
    };

    el.lists.appendChild(li);
  });
}

function renderItems() {
  el.items.innerHTML = "";
  state.items.forEach((item) => {
    const li = document.createElement("li");
    li.innerHTML = `
      <div class="item-title">${item.done ? "✅" : "⬜"} ${item.title}</div>
      <div class="item-meta">${item.description || "без описания"}</div>
      <div class="item-actions">
        <button class="small secondary" data-act="toggle">${item.done ? "Вернуть" : "Готово"}</button>
        <button class="small secondary" data-act="edit">Изменить</button>
        <button class="small danger" data-act="delete">Удалить</button>
      </div>
    `;

    li.querySelector('[data-act="toggle"]').onclick = async () => {
      await api(`/lists/${state.selectedListId}/items/${item.id}/`, {
        method: "PATCH",
        body: JSON.stringify({ done: !item.done }),
      });
      await loadItems();
    };

    li.querySelector('[data-act="edit"]').onclick = async () => {
      const title = prompt("Название", item.title);
      if (title === null) return;
      const description = prompt("Описание", item.description || "");
      if (description === null) return;
      await api(`/lists/${state.selectedListId}/items/${item.id}/`, {
        method: "PATCH",
        body: JSON.stringify({ title, description }),
      });
      showToast("Задача обновлена");
      await loadItems();
    };

    li.querySelector('[data-act="delete"]').onclick = async () => {
      if (!confirm("Удалить задачу?")) return;
      await api(`/lists/${state.selectedListId}/items/${item.id}/`, { method: "DELETE" });
      showToast("Задача удалена");
      await loadItems();
    };

    el.items.appendChild(li);
  });
}

async function loadLists() {
  if (!state.token) return;
  state.lists = await api("/lists/");
  if (state.selectedListId && !state.lists.find((x) => x.id === state.selectedListId)) {
    state.selectedListId = null;
    state.items = [];
  }
  renderLists();
}

async function loadItems() {
  if (!state.token || !state.selectedListId) {
    state.items = [];
    renderItems();
    return;
  }
  state.items = await api(`/lists/${state.selectedListId}/items/`);
  renderItems();
}

function switchAuthTab(signIn) {
  el.tabSignIn.classList.toggle("active", signIn);
  el.tabSignUp.classList.toggle("active", !signIn);
  el.signInForm.classList.toggle("hidden", !signIn);
  el.signUpForm.classList.toggle("hidden", signIn);
}

function bindEvents() {
  el.apiBase.value = state.apiBase;
  el.saveApiBase.onclick = () => {
    state.apiBase = el.apiBase.value.trim().replace(/\/$/, "");
    localStorage.setItem("todo_api_base", state.apiBase);
    showToast("API URL сохранен");
  };

  el.tabSignIn.onclick = () => switchAuthTab(true);
  el.tabSignUp.onclick = () => switchAuthTab(false);

  el.signInForm.onsubmit = async (e) => {
    e.preventDefault();
    try {
      const res = await api("/auth/sign-in", {
        method: "POST",
        body: JSON.stringify({
          username: el.signInUsername.value.trim(),
          password: el.signInPassword.value,
        }),
      });
      setToken(res.token);
      showToast("Успешный вход");
      await loadLists();
    } catch (err) {
      showToast(err.message);
    }
  };

  el.signUpForm.onsubmit = async (e) => {
    e.preventDefault();
    try {
      const res = await api("/auth/sign-up", {
        method: "POST",
        body: JSON.stringify({
          name: el.signUpName.value.trim(),
          username: el.signUpUsername.value.trim(),
          password: el.signUpPassword.value,
        }),
      });
      setToken(res.token);
      showToast("Пользователь создан");
      switchAuthTab(true);
      await loadLists();
    } catch (err) {
      showToast(err.message);
    }
  };

  el.logoutBtn.onclick = () => {
    setToken("");
    state.lists = [];
    state.items = [];
    state.selectedListId = null;
    renderLists();
    renderItems();
    showToast("Вы вышли");
  };

  el.refreshLists.onclick = loadLists;
  el.refreshItems.onclick = loadItems;

  el.createListForm.onsubmit = async (e) => {
    e.preventDefault();
    try {
      await api("/lists/", {
        method: "POST",
        body: JSON.stringify({
          title: el.newListTitle.value.trim(),
          description: el.newListDescription.value.trim(),
        }),
      });
      el.newListTitle.value = "";
      el.newListDescription.value = "";
      showToast("Список создан");
      await loadLists();
    } catch (err) {
      showToast(err.message);
    }
  };

  el.createItemForm.onsubmit = async (e) => {
    e.preventDefault();
    if (!state.selectedListId) {
      showToast("Сначала выбери список");
      return;
    }

    try {
      await api(`/lists/${state.selectedListId}/items/`, {
        method: "POST",
        body: JSON.stringify({
          title: el.newItemTitle.value.trim(),
          description: el.newItemDescription.value.trim(),
        }),
      });
      el.newItemTitle.value = "";
      el.newItemDescription.value = "";
      showToast("Задача создана");
      await loadItems();
    } catch (err) {
      showToast(err.message);
    }
  };
}

async function init() {
  bindEvents();
  setToken(state.token);
  switchAuthTab(true);

  if (state.token) {
    try {
      await loadLists();
    } catch (err) {
      showToast(err.message);
      setToken("");
    }
  }
}

init();
