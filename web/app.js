// app.js — клиентская логика мини-приложения FitTrack.
// Работает внутри Telegram Mini App и в обычном браузере (демо-режим).

"use strict";

// --- Интеграция с Telegram --------------------------------------------------

const tg = window.Telegram && window.Telegram.WebApp ? window.Telegram.WebApp : null;
if (tg) {
  tg.ready();
  tg.expand();
  try { tg.setHeaderColor("#ffffff"); } catch (e) { /* не поддерживается */ }
}

// Заголовок с initData для проверки подписи на сервере.
function authHeaders() {
  if (tg && tg.initData) {
    return { "X-Telegram-Init-Data": tg.initData };
  }
  return {}; // браузерный демо-режим — сервер применит гостевой профиль
}

// --- HTTP-помощники ---------------------------------------------------------

async function apiGet(path) {
  const res = await fetch(path, { headers: authHeaders() });
  if (!res.ok) throw new Error((await res.json()).error || res.status);
  return res.json();
}

async function apiPost(path, body, method = "POST") {
  const res = await fetch(path, {
    method,
    headers: { "Content-Type": "application/json", ...authHeaders() },
    body: JSON.stringify(body),
  });
  const data = await res.json();
  if (!res.ok) throw new Error(data.error || res.status);
  return data;
}

// --- Навигация --------------------------------------------------------------

const screens = document.querySelectorAll(".screen");
const tabs = document.querySelectorAll(".tabbar button");

function showScreen(name) {
  screens.forEach((s) => { s.hidden = s.id !== "screen-" + name; });
  tabs.forEach((b) => b.classList.toggle("active", b.dataset.screen === name));
  if (name === "home") loadHome();
  if (name === "catalog") loadCatalog();
  if (name === "challenges") loadChallenges();
  if (name === "reminders") loadReminders();
  if (name === "chat") loadChat();
  if (name === "profile") loadProfile();
}

document.getElementById("tabbar").addEventListener("click", (e) => {
  const btn = e.target.closest("button[data-screen]");
  if (btn) showScreen(btn.dataset.screen);
});

// Быстрые карточки на главной.
document.getElementById("miniCatalog").onclick = () => showScreen("catalog");
document.getElementById("miniBmi").onclick = () => showScreen("bmi");
document.getElementById("miniChallenges").onclick = () => showScreen("challenges");
document.getElementById("miniCoach").onclick = () => showScreen("coach");

// --- Главная: шагомер (заглушка) и сводка недели -----------------------------

const fmt = (n) => n.toLocaleString("ru-RU");

async function loadHome() {
  try {
    const stats = await apiGet("/api/steps");
    const pct = Math.min(100, Math.round(stats.total / 7 / stats.goal * 100));
    document.getElementById("stepsRing").style.setProperty("--pct", pct);
    document.getElementById("stepsToday").textContent = fmt(stats.days[6].steps);
    document.getElementById("stepsKm").textContent = String(stats.distanceKm).replace(".", ",");
    document.getElementById("stepsKcal").textContent = fmt(stats.kcal);
    document.getElementById("stepsStreak").textContent = stats.streak;
    document.getElementById("stepsDate").value = new Date().toISOString().slice(0, 10);
    renderWeekChart(stats);
    loadLevelChip();
  } catch (e) {
    console.warn("шагомер:", e.message);
  }
}

// Шагомер работает в режиме заглушки: значения вводятся вручную.
// Точка будущей интеграции: navigator.hardware/сенсоры телефона или Health API.
document.getElementById("stepsForm").addEventListener("submit", async (e) => {
  e.preventDefault();
  try {
    await apiPost("/api/steps", {
      date: document.getElementById("stepsDate").value,
      steps: Number(document.getElementById("stepsInput").value),
    });
    document.getElementById("stepsInput").value = "";
    loadHome();
  } catch (err) {
    alert(err.message);
  }
});

function renderWeekChart(stats) {
  const chart = document.getElementById("weekChart");
  chart.innerHTML = "";
  const max = Math.max(stats.goal, ...stats.days.map((d) => d.steps), 1);
  stats.days.forEach((d) => {
    const bar = document.createElement("div");
    bar.className = "bar" + (d.steps >= stats.goal ? " done" : "");
    bar.style.height = Math.round(d.steps / max * 100) + "%";
    bar.innerHTML = "<small>" + d.date.slice(8) + "</small>";
    chart.appendChild(bar);
  });
}

// --- Каталог ----------------------------------------------------------------

const WEEKDAYS = ["вс", "пн", "вт", "ср", "чт", "пт", "сб"];

async function loadCatalog() {
  try {
    const q = document.getElementById("catalogSearch").value;
    const cat = document.getElementById("catalogCategory").value;
    const data = await apiGet("/api/catalog?q=" + encodeURIComponent(q) + "&category=" + encodeURIComponent(cat));
    const select = document.getElementById("catalogCategory");
    if (select.options.length <= 1) {
      data.categories.forEach((c) => {
        const opt = document.createElement("option");
        opt.value = c; opt.textContent = c[0].toUpperCase() + c.slice(1);
        select.appendChild(opt);
      });
    }
    const list = document.getElementById("productList");
    list.innerHTML = "";
    if (!data.products.length) {
      list.innerHTML = '<p class="empty">Ничего не найдено</p>';
      return;
    }
    data.products.forEach((p) => {
      const el = document.createElement("div");
      el.className = "product";
      el.innerHTML =
        "<div><b>" + p.name + "</b>" +
        '<span class="tag">' + p.category + "</span> " +
        (p.inStock ? "" : '<span class="tag out">нет в наличии</span>') +
        "<small>" + p.desc + "</small></div>" +
        '<div class="price">' + fmt(p.price) + " ₽<small>★ " + p.rating.toFixed(1) + "</small></div>";
      list.appendChild(el);
    });
  } catch (e) {
    console.warn("каталог:", e.message);
  }
}

document.getElementById("catalogSearch").addEventListener("input", debounce(loadCatalog, 300));
document.getElementById("catalogCategory").addEventListener("change", loadCatalog);

// --- ИМТ --------------------------------------------------------------------

document.getElementById("bmiForm").addEventListener("submit", async (e) => {
  e.preventDefault();
  try {
    const res = await apiPost("/api/bmi", {
      weight: Number(document.getElementById("bmiWeight").value),
      height: Number(document.getElementById("bmiHeight").value),
    });
    document.getElementById("bmiResult").hidden = false;
    document.getElementById("bmiValue").textContent = String(res.value).replace(".", ",");
    document.getElementById("bmiCategory").textContent = res.category;
    document.getElementById("bmiAdvice").textContent = res.advice;
  } catch (err) {
    alert(err.message);
  }
});

// --- Челленджи --------------------------------------------------------------

async function loadChallenges() {
  try {
    const list = await apiGet("/api/challenges");
    const box = document.getElementById("challengeList");
    box.innerHTML = "";
    if (!list.length) {
      box.innerHTML = '<p class="empty">Челленджей пока нет</p>';
      return;
    }
    list.forEach((c) => {
      const el = document.createElement("div");
      el.className = "challenge";
      el.innerHTML =
        "<header><b>" + c.title + "</b>" +
        (c.joined ? '<span class="tag">вы участвуете</span>' : '<button class="btn" data-code="' + c.code + '">Участвовать</button>') +
        "</header>" +
        "<small>" + c.desc + "</small>" +
        '<div class="progress-track"><i style="width:' + Math.min(100, c.pace.donePct) + '%"></i></div>' +
        '<div class="meta"><span>' + Math.round(c.pace.donePct) + "% цели</span>" +
        "<span>осталось " + fmt(c.pace.remaining) + " шагов</span>" +
        "<span>темп " + fmt(c.pace.perDay) + "/день (" + c.pace.daysLeft + " дн.)</span>" +
        "<span>" + c.members + " участн.</span></div>";
      box.appendChild(el);
    });
    box.querySelectorAll("button[data-code]").forEach((b) => {
      b.onclick = async () => {
        try {
          await apiPost("/api/challenges/" + b.dataset.code + "/join");
          loadChallenges();
        } catch (e) { alert(e.message); }
      };
    });
  } catch (e) {
    console.warn("челленджи:", e.message);
  }
}

// --- Напоминания ------------------------------------------------------------

const dayPicker = document.getElementById("weekdayPicker");
[1, 2, 3, 4, 5, 6, 0].forEach((d) => {
  const label = document.createElement("label");
  label.innerHTML = '<input type="checkbox" value="' + d + '">' + WEEKDAYS[d];
  dayPicker.appendChild(label);
});

async function loadReminders() {
  try {
    const list = await apiGet("/api/reminders");
    const box = document.getElementById("reminderList");
    box.innerHTML = "";
    if (!list.length) {
      box.innerHTML = '<p class="empty">Напоминаний пока нет</p>';
      return;
    }
    list.forEach((r) => {
      const el = document.createElement("div");
      el.className = "reminder";
      el.innerHTML =
        "<div><b>" + escapeHTML(r.title) + "</b><small>" +
        r.days.map((d) => WEEKDAYS[d]).join(", ") + "</small></div>" +
        '<div><span class="time">' + r.time + '</span> <button title="Удалить">✕</button></div>';
      el.querySelector("button").onclick = async () => {
        try { await apiPost("/api/reminders/" + r.id, null, "DELETE"); loadReminders(); }
        catch (e) { alert(e.message); }
      };
      box.appendChild(el);
    });
  } catch (e) {
    console.warn("напоминания:", e.message);
  }
}

document.getElementById("reminderForm").addEventListener("submit", async (e) => {
  e.preventDefault();
  const days = [...dayPicker.querySelectorAll("input:checked")].map((i) => Number(i.value));
  try {
    await apiPost("/api/reminders", {
      title: document.getElementById("remTitle").value,
      time: document.getElementById("remTime").value,
      days,
    });
    e.target.reset();
    dayPicker.querySelectorAll("input").forEach((i) => { i.checked = false; });
    loadReminders();
  } catch (err) {
    alert(err.message);
  }
});

// --- Чат --------------------------------------------------------------------

async function loadChat() {
  try {
    const messages = await apiGet("/api/chat");
    const box = document.getElementById("chatBox");
    box.innerHTML = "";
    if (!messages.length) {
      box.innerHTML = '<p class="empty">Сообщений пока нет — напишите первым</p>';
      return;
    }
    messages.forEach((m) => box.appendChild(bubble(m, m.me)));
    box.scrollTop = box.scrollHeight;
  } catch (e) {
    console.warn("чат:", e.message);
  }
}

function bubble(m, mine) {
  const el = document.createElement("div");
  el.className = "msg" + (mine ? " me" : "");
  el.innerHTML = "<b>" + escapeHTML(m.user) + "</b>" + escapeHTML(m.text);
  return el;
}

document.getElementById("chatForm").addEventListener("submit", async (e) => {
  e.preventDefault();
  const input = document.getElementById("chatInput");
  const text = input.value.trim();
  if (!text) return;
  try {
    await apiPost("/api/chat", { text });
    input.value = "";
    loadChat();
  } catch (err) {
    alert(err.message);
  }
});

// --- ИИ-тренер --------------------------------------------------------------

document.getElementById("coachForm").addEventListener("submit", async (e) => {
  e.preventDefault();
  const input = document.getElementById("coachInput");
  const text = input.value.trim();
  if (!text) return;
  const box = document.getElementById("coachBox");
  box.appendChild(Object.assign(document.createElement("div"), { className: "msg me", textContent: text }));
  input.value = "";
  box.scrollTop = box.scrollHeight;
  try {
    const res = await apiPost("/api/ai/chat", { message: text });
    const el = document.createElement("div");
    el.className = "msg bot";
    el.innerHTML = "<b>Тренер" + (res.source === "demo" ? " (демо-режим)" : "") + "</b>" + escapeHTML(res.reply);
    box.appendChild(el);
    box.scrollTop = box.scrollHeight;
  } catch (err) {
    alert(err.message);
  }
});

// --- Профиль и геймификация -------------------------------------------------

async function loadProfile() {
  try {
    const data = await apiGet("/api/profile");
    const p = data.profile;
    document.getElementById("profName").value = p.name || "";
    document.getElementById("profWeight").value = p.weight || "";
    document.getElementById("profHeight").value = p.height || "";
    document.getElementById("profGoal").value = p.goal || 8000;
    document.getElementById("profPoints").textContent = fmt(data.progress.points || 0);
    document.getElementById("profLevel").textContent = data.progress.level || 1;
    const box = document.getElementById("achievementList");
    box.innerHTML = "";
    const got = data.progress.achievements || [];
    (data.catalog || []).forEach((a) => {
      const el = document.createElement("span");
      el.className = "ach" + (got.includes(a.code) ? "" : " locked");
      el.textContent = (got.includes(a.code) ? "✓ " : "🔒 ") + a.name;
      box.appendChild(el);
    });
  } catch (e) {
    console.warn("профиль:", e.message);
  }
}

document.getElementById("profileForm").addEventListener("submit", async (e) => {
  e.preventDefault();
  try {
    await apiPost("/api/profile", {
      name: document.getElementById("profName").value,
      weight: Number(document.getElementById("profWeight").value),
      height: Number(document.getElementById("profHeight").value),
      goal: Number(document.getElementById("profGoal").value),
    });
    loadLevelChip();
  } catch (err) {
    alert(err.message);
  }
});

async function loadLevelChip() {
  try {
    const data = await apiGet("/api/profile");
    document.getElementById("levelNum").textContent = "Ур. " + (data.progress.level || 1);
    document.getElementById("levelFill").style.width = (data.progress.levelProgress || 0) + "%";
  } catch (e) { /* не критично */ }
}

// --- Служебные функции -------------------------------------------------------

function escapeHTML(s) {
  return String(s == null ? "" : s)
    .replaceAll("&", "&amp;").replaceAll("<", "&lt;").replaceAll(">", "&gt;")
    .replaceAll('"', "&quot;").replaceAll("'", "&#39;");
}

function debounce(fn, ms) {
  let timer;
  return (...args) => {
    clearTimeout(timer);
    timer = setTimeout(() => fn(...args), ms);
  };
}

// Стартовый экран.
showScreen("home");
