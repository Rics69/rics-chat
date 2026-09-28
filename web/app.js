"use strict";

const TOKEN_KEY = "rics-chat-token";
const PAGE_SIZE = 30;

const $ = (id) => document.getElementById(id);

const state = {
  token: localStorage.getItem(TOKEN_KEY),
  me: null,
  peer: null,
  messages: new Map(),
  nextBeforeId: null,
  loadingOlder: false,
  dialogs: new Map(),
  ws: null,
  wsRetry: 0,
  wsStopped: false,
};

// ---------- API ----------

async function api(method, path, body) {
  const headers = { "Content-Type": "application/json" };
  if (state.token) {
    headers.Authorization = `Bearer ${state.token}`;
  }

  const resp = await fetch(path, {
    method,
    headers,
    body: body ? JSON.stringify(body) : undefined,
  });
  const data = await resp.json().catch(() => ({}));

  if (!resp.ok) {
    if (resp.status === 401 && state.token) {
      logout();
    }
    throw new Error(data?.error?.message || `HTTP ${resp.status}`);
  }

  return data;
}

// id приходят строками (int64 в JSON), сравниваем через BigInt
const cmpId = (a, b) => (BigInt(a) < BigInt(b) ? -1 : BigInt(a) > BigInt(b) ? 1 : 0);

function formatTime(iso) {
  const date = new Date(iso);
  const now = new Date();
  if (date.toDateString() === now.toDateString()) {
    return date.toLocaleTimeString("ru-RU", { hour: "2-digit", minute: "2-digit" });
  }
  return date.toLocaleDateString("ru-RU", { day: "2-digit", month: "2-digit" });
}

function el(tag, className, text) {
  const node = document.createElement(tag);
  if (className) node.className = className;
  // только textContent, никакого innerHTML с пользовательскими данными — иначе XSS
  if (text !== undefined) node.textContent = text;
  return node;
}

// ---------- auth ----------

$("auth-form").addEventListener("submit", async (e) => {
  e.preventDefault();
  const action = e.submitter?.dataset.action || "login";
  $("auth-error").textContent = "";

  try {
    const data = await api("POST", `/api/v1/auth/${action}`, {
      login: $("auth-login").value,
      password: $("auth-password").value,
    });
    state.token = data.access_token;
    localStorage.setItem(TOKEN_KEY, state.token);
    await start(data.user);
  } catch (err) {
    $("auth-error").textContent = err.message;
  }
});

$("logout").addEventListener("click", logout);

function logout() {
  stopWS();
  state.dialogs.clear();
  closeDialog();
  state.token = null;
  state.me = null;
  localStorage.removeItem(TOKEN_KEY);
  $("chat-screen").hidden = true;
  $("auth-screen").hidden = false;
}

async function start(me) {
  state.me = me;
  $("me-login").textContent = me.login;
  $("auth-screen").hidden = true;
  $("chat-screen").hidden = false;
  connectWS();
  await loadDialogs();
}

// ---------- dialogs ----------

async function loadDialogs() {
  const data = await api("GET", "/api/v1/dialogs");
  const unread = new Map([...state.dialogs].map(([id, d]) => [id, d.unread]));

  state.dialogs.clear();
  for (const dialog of data.dialogs) {
    state.dialogs.set(dialog.peer.id, { ...dialog, unread: unread.get(dialog.peer.id) || 0 });
  }
  renderDialogs();
}

function renderDialogs() {
  const list = $("dialogs");
  list.replaceChildren();

  const dialogs = [...state.dialogs.values()].sort((a, b) =>
    cmpId(b.last_message.id, a.last_message.id),
  );

  if (dialogs.length === 0) {
    list.append(el("li", "hint", "Пока нет диалогов — найдите кого-нибудь через поиск"));
    return;
  }

  for (const dialog of dialogs) {
    const li = el("li");
    if (state.peer?.id === dialog.peer.id) li.classList.add("active");

    const mine = dialog.last_message.sender_id === state.me.id;
    li.append(
      el("span", "login", dialog.peer.login),
      el("span", "time", formatTime(dialog.last_message.created_at)),
      el("span", "preview", (mine ? "Вы: " : "") + dialog.last_message.text),
    );
    if (dialog.unread > 0) li.append(el("span", "badge", String(dialog.unread)));

    li.addEventListener("click", () => openDialog(dialog.peer));
    list.append(li);
  }
}

// ---------- search ----------

let searchTimer;
$("search").addEventListener("input", () => {
  clearTimeout(searchTimer);
  // debounce: не дёргаем сервер на каждую букву
  searchTimer = setTimeout(search, 300);
});

async function search() {
  const query = $("search").value.trim();
  const list = $("search-results");

  if (!query) {
    list.hidden = true;
    $("dialogs").hidden = false;
    return;
  }

  list.hidden = false;
  $("dialogs").hidden = true;

  try {
    const data = await api("GET", `/api/v1/users/search?query=${encodeURIComponent(query)}`);
    if ($("search").value.trim() !== query) return;

    list.replaceChildren();
    if (data.users.length === 0) {
      list.append(el("li", "hint", "Никого не нашли"));
    }
    for (const user of data.users) {
      const li = el("li");
      li.append(el("span", "login", user.login));
      li.addEventListener("click", () => {
        $("search").value = "";
        search();
        openDialog(user);
      });
      list.append(li);
    }
  } catch (err) {
    list.replaceChildren(el("li", "hint", err.message));
  }
}

// ---------- conversation ----------

async function openDialog(peer) {
  state.peer = { id: peer.id, login: peer.login };
  state.messages.clear();
  state.nextBeforeId = null;

  const dialog = state.dialogs.get(peer.id);
  if (dialog) dialog.unread = 0;

  $("peer-login").textContent = peer.login;
  $("empty-state").hidden = true;
  $("conversation").hidden = false;
  $("chat-screen").classList.add("has-peer");
  $("messages").replaceChildren();
  renderDialogs();

  await loadLatestMessages();
  $("send-text").focus();
}

function closeDialog() {
  state.peer = null;
  state.messages.clear();
  $("conversation").hidden = true;
  $("empty-state").hidden = false;
  $("chat-screen").classList.remove("has-peer");
  renderDialogs();
}

$("back").addEventListener("click", closeDialog);

async function loadLatestMessages() {
  const peerId = state.peer.id;
  const data = await api("GET", `/api/v1/dialogs/${peerId}/messages?limit=${PAGE_SIZE}`);

  // пока ждали ответ, пользователь мог открыть другой диалог
  if (state.peer?.id !== peerId) return;

  addMessages(data.messages);
  if (state.nextBeforeId === null) {
    state.nextBeforeId = data.next_before_id ?? null;
  }
  renderMessages({ scrollToBottom: true });
}

async function loadOlderMessages() {
  if (state.loadingOlder || state.nextBeforeId === null) return;
  state.loadingOlder = true;

  const peerId = state.peer.id;
  try {
    const data = await api(
      "GET",
      `/api/v1/dialogs/${peerId}/messages?limit=${PAGE_SIZE}&before_id=${state.nextBeforeId}`,
    );
    if (state.peer?.id !== peerId) return;

    addMessages(data.messages);
    state.nextBeforeId = data.next_before_id ?? null;
    renderMessages({ keepScroll: true });
  } finally {
    state.loadingOlder = false;
  }
}

$("messages").addEventListener("scroll", () => {
  if ($("messages").scrollTop < 80) loadOlderMessages();
});

// Map по id: одно и то же сообщение приходит и в ответе SendMessage, и из сокета — дубль не появится
function addMessages(messages) {
  for (const message of messages) {
    state.messages.set(message.id, message);
  }
}

function renderMessages({ scrollToBottom = false, keepScroll = false } = {}) {
  const box = $("messages");
  const nearBottom = box.scrollHeight - box.scrollTop - box.clientHeight < 80;
  const prevHeight = box.scrollHeight;

  const messages = [...state.messages.values()].sort((a, b) => cmpId(a.id, b.id));
  box.replaceChildren(
    ...messages.map((message) => {
      const out = message.sender_id === state.me.id;
      const node = el("div", out ? "message out" : "message", message.text);
      node.append(el("span", "time", formatTime(message.created_at)));
      return node;
    }),
  );

  if (keepScroll) {
    box.scrollTop += box.scrollHeight - prevHeight;
  } else if (scrollToBottom || nearBottom) {
    box.scrollTop = box.scrollHeight;
  }
}

$("send-form").addEventListener("submit", async (e) => {
  e.preventDefault();
  const input = $("send-text");
  const text = input.value;
  if (!text.trim() || !state.peer) return;

  input.value = "";
  try {
    const data = await api("POST", "/api/v1/messages", {
      recipient_id: state.peer.id,
      text,
    });
    onNewMessage(data.message);
  } catch (err) {
    input.value = text;
    alert(err.message);
  }
});

// ---------- realtime ----------

function connectWS() {
  state.wsStopped = false;
  const proto = location.protocol === "https:" ? "wss" : "ws";
  const ws = new WebSocket(`${proto}://${location.host}/ws?token=${encodeURIComponent(state.token)}`);
  state.ws = ws;

  ws.onopen = () => {
    const reconnected = state.wsRetry > 0;
    state.wsRetry = 0;
    $("ws-status").classList.add("online");

    // пока были офлайн, события терялись — догружаем из истории
    if (reconnected) {
      loadDialogs().catch(() => {});
      if (state.peer) loadLatestMessages().catch(() => {});
    }
  };

  ws.onmessage = (e) => {
    const event = JSON.parse(e.data);
    if (event.new_message) onNewMessage(event.new_message);
  };

  ws.onclose = (e) => {
    $("ws-status").classList.remove("online");
    if (state.wsStopped || state.ws !== ws) return;

    // 4001 — сервер закрыл сокет, потому что истёк токен
    if (e.code === 4001) {
      logout();
      return;
    }

    // экспоненциальная задержка: 1, 2, 4, 8, 10, 10... секунд
    const delay = Math.min(1000 * 2 ** state.wsRetry, 10000);
    state.wsRetry++;
    setTimeout(() => {
      if (!state.wsStopped && state.ws === ws) connectWS();
    }, delay);
  };
}

function stopWS() {
  state.wsStopped = true;
  state.ws?.close();
  state.ws = null;
}

function onNewMessage(message) {
  const peerId = message.sender_id === state.me.id ? message.recipient_id : message.sender_id;
  const isOpen = state.peer?.id === peerId;

  if (isOpen) {
    const isNew = !state.messages.has(message.id);
    addMessages([message]);
    if (isNew) renderMessages();
  }

  const dialog = state.dialogs.get(peerId);
  if (!dialog) {
    // новый собеседник: логина у нас нет, проще перезапросить список
    loadDialogs()
      .then(() => {
        const created = state.dialogs.get(peerId);
        if (created && !isOpen && message.sender_id !== state.me.id) {
          created.unread++;
          renderDialogs();
        }
      })
      .catch(() => {});
    return;
  }

  if (cmpId(message.id, dialog.last_message.id) > 0) {
    dialog.last_message = message;
    if (!isOpen && message.sender_id !== state.me.id) dialog.unread++;
  }
  renderDialogs();
}

// ---------- boot ----------

(async function boot() {
  if (!state.token) {
    $("auth-screen").hidden = false;
    return;
  }

  try {
    const data = await api("GET", "/api/v1/users/me");
    await start(data.user);
  } catch {
    logout();
  }
})();
