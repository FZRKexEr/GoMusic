const form = document.querySelector("#parseForm");
const shareInput = document.querySelector("#shareInput");
const parseButton = document.querySelector("#parseButton");
const clearButton = document.querySelector("#clearButton");
const reverseOrder = document.querySelector("#reverseOrder");
const statusText = document.querySelector("#statusText");
const playlistName = document.querySelector("#playlistName");
const songCount = document.querySelector("#songCount");
const emptyState = document.querySelector("#emptyState");
const resultToolbar = document.querySelector("#resultToolbar");
const songList = document.querySelector("#songList");
const copySongsButton = document.querySelector("#copySongsButton");
const copyJsonButton = document.querySelector("#copyJsonButton");

let currentResult = null;

form.addEventListener("submit", async (event) => {
  event.preventDefault();
  const text = shareInput.value.trim();

  if (!text) {
    setStatus("请输入分享内容", "error");
    shareInput.focus();
    return;
  }

  parseButton.disabled = true;
  setStatus("解析中");

  try {
    const result = await requestSongList(text);
    if (result.code !== 0) {
      throw new Error(result.msg || "解析失败");
    }

    currentResult = result.data;
    renderResult(currentResult);
    setStatus("解析完成", "toast");
  } catch (error) {
    currentResult = null;
    renderEmpty(error.message || "解析失败", "解析失败");
    setStatus(error.message || "解析失败", "error");
  } finally {
    parseButton.disabled = false;
  }
});

clearButton.addEventListener("click", () => {
  shareInput.value = "";
  currentResult = null;
  renderEmpty("暂无结果");
  setStatus("就绪");
  shareInput.focus();
});

copySongsButton.addEventListener("click", async () => {
  if (!currentResult) return;
  await copyText(currentResult.songs.join("\n"));
  setStatus("已复制歌曲", "toast");
});

copyJsonButton.addEventListener("click", async () => {
  if (!currentResult) return;
  await copyText(JSON.stringify(currentResult, null, 2));
  setStatus("已复制 JSON", "toast");
});

async function requestSongList(text) {
  const detailed = form.elements.detailed.value;
  const format = form.elements.format.value;
  const order = reverseOrder.checked ? "reverse" : "";
  const params = new URLSearchParams({ detailed, format });
  if (order) params.set("order", order);

  const body = new URLSearchParams({ url: text });
  const response = await fetch(`/songlist?${params.toString()}`, {
    method: "POST",
    headers: {
      "Content-Type": "application/x-www-form-urlencoded",
    },
    body,
  });

  const payload = await response.json().catch(() => ({
    code: response.status,
    msg: "响应格式错误",
    data: null,
  }));

  if (!response.ok && payload.code === 0) {
    payload.code = response.status;
  }
  return payload;
}

function renderResult(data) {
  const songs = Array.isArray(data.songs) ? data.songs : [];
  playlistName.textContent = data.name || "未命名歌单";
  songCount.textContent = `${data.songs_count || songs.length} 首`;
  songList.innerHTML = "";

  const fragment = document.createDocumentFragment();
  songs.forEach((song) => {
    const item = document.createElement("li");
    const title = document.createElement("span");
    title.className = "song-title";
    title.textContent = song;
    item.append(title);
    fragment.append(item);
  });
  songList.append(fragment);

  emptyState.hidden = songs.length > 0;
  resultToolbar.hidden = songs.length === 0;
}

function renderEmpty(message, title = "等待输入") {
  playlistName.textContent = title;
  songCount.textContent = "0 首";
  songList.innerHTML = "";
  resultToolbar.hidden = true;
  emptyState.hidden = false;
  emptyState.querySelector("p").textContent = message;
}

function setStatus(message, className = "") {
  statusText.textContent = message;
  statusText.className = `status ${className}`.trim();
}

async function copyText(text) {
  if (navigator.clipboard?.writeText) {
    await navigator.clipboard.writeText(text);
    return;
  }

  const textarea = document.createElement("textarea");
  textarea.value = text;
  textarea.setAttribute("readonly", "");
  textarea.style.position = "fixed";
  textarea.style.opacity = "0";
  document.body.append(textarea);
  textarea.select();
  document.execCommand("copy");
  textarea.remove();
}
