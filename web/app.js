const transcript = document.querySelector("#transcript");
const empty = document.querySelector("#empty");
const promptEl = document.querySelector("#prompt");
const contextEl = document.querySelector("#context");
const form = document.querySelector("#composer");
const sendBtn = document.querySelector("#send");
const providerEl = document.querySelector("#provider");
const messages = [];

function addBubble(role, text) {
  if (empty) {
    empty.remove();
  }
  const item = document.createElement("article");
  item.className = "bubble " + role;
  const who = document.createElement("p");
  who.className = "who";
  who.textContent = role === "user" ? "You" : "Go-VS-Copilot";
  const body = document.createElement("pre");
  body.textContent = text;
  item.append(who, body);
  transcript.append(item);
  item.scrollIntoView({ block: "nearest" });
}

async function loadHealth() {
  const response = await fetch("/api/health");
  const data = await response.json();
  if (!response.ok) {
    providerEl.textContent = "Unavailable";
    return;
  }
  providerEl.textContent = data.provider === "stub"
    ? "Local stub"
    : data.provider + " · " + data.model;
}

form.addEventListener("submit", async (event) => {
  event.preventDefault();
  const prompt = promptEl.value.trim();
  if (!prompt) {
    promptEl.focus();
    return;
  }
  const context = contextEl.value;
  const history = messages.slice();
  addBubble("user", prompt);
  promptEl.value = "";
  sendBtn.disabled = true;
  try {
    const response = await fetch("/api/complete", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ prompt, context, messages: history }),
    });
    const data = await response.json().catch(() => ({}));
    if (!response.ok) {
      throw new Error(data.error || "Request failed");
    }
    addBubble("assistant", data.reply);
    messages.push({ role: "user", content: prompt });
    messages.push({ role: "assistant", content: data.reply });
  } catch (err) {
    addBubble("assistant", err instanceof Error ? err.message : "Request failed");
  } finally {
    sendBtn.disabled = false;
    promptEl.focus();
  }
});

promptEl.addEventListener("keydown", (event) => {
  if (event.key === "Enter" && (event.metaKey || event.ctrlKey)) {
    event.preventDefault();
    form.requestSubmit();
  }
});

loadHealth().catch(() => {
  providerEl.textContent = "Offline";
});

if ("serviceWorker" in navigator) {
  navigator.serviceWorker.register("/sw.js").catch(() => {});
}
