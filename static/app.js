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
const copyPanel = document.querySelector("#copyPanel");
const copyBuffer = document.querySelector("#copyBuffer");

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
  hideCopyPanel();
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
  hideCopyPanel();
  renderEmpty("暂无结果");
  setStatus("就绪");
  shareInput.focus();
});

copySongsButton.addEventListener("click", async () => {
  if (!currentResult) {
    setStatus("暂无可复制内容", "error");
    return;
  }

  await copyResultText(currentResult.songs.join("\n"), "已复制歌曲");
});

copyJsonButton.addEventListener("click", async () => {
  if (!currentResult) {
    setStatus("暂无可复制内容", "error");
    return;
  }

  await copyResultText(JSON.stringify(currentResult, null, 2), "已复制 JSON");
});

async function requestSongList(text) {
  const clean = form.elements.clean.value === "true";
  const format = form.elements.format.value;
  const response = await fetch("/songlist", {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify({
      url: text,
      clean,
      format,
      reverse: reverseOrder.checked,
    }),
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
  setCopyPanelText(songs.join("\n"));
  setElementVisible(copyPanel, songs.length > 0);

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

  setElementVisible(emptyState, songs.length === 0);
  setElementVisible(songList, songs.length > 0);
  setElementVisible(resultToolbar, songs.length > 0);
}

function renderEmpty(message, title = "等待输入") {
  playlistName.textContent = title;
  songCount.textContent = "0 首";
  songList.innerHTML = "";
  hideCopyPanel();
  setElementVisible(songList, false);
  setElementVisible(resultToolbar, false);
  setElementVisible(emptyState, true);
  emptyState.querySelector("p").textContent = message;
}

function setStatus(message, className = "") {
  statusText.textContent = message;
  statusText.className = `status ${className}`.trim();
}

function setElementVisible(element, visible) {
  element.hidden = !visible;
  element.style.display = visible ? "" : "none";
  element.setAttribute("aria-hidden", visible ? "false" : "true");
}

async function copyText(text) {
  if (copyTextWithCommand(text)) return;
  if (copyTextWithSelection(text)) return;

  if (window.isSecureContext && navigator.clipboard?.writeText) {
    try {
      await navigator.clipboard.writeText(text);
      return;
    } catch (error) {
      throw error;
    }
  }

  throw new Error("复制失败");
}

async function copyResultText(text, successMessage) {
  setCopyPanelText(text);

  try {
    await copyText(text);
    setStatus(successMessage, "toast");
  } catch {
    selectCopyPanel();
    setStatus("内容已选中，可手动复制", "toast");
  }
}

function setCopyPanelText(text) {
  copyBuffer.value = text;
}

function selectCopyPanel() {
  setElementVisible(copyPanel, true);
  copyBuffer.focus();
  copyBuffer.select();
  copyBuffer.setSelectionRange(0, copyBuffer.value.length);
}

function hideCopyPanel() {
  copyBuffer.value = "";
  setElementVisible(copyPanel, false);
}

function copyTextWithCommand(text) {
  let copied = false;
  const onCopy = (event) => {
    event.clipboardData.setData("text/plain", text);
    event.preventDefault();
    copied = true;
  };

  document.addEventListener("copy", onCopy);
  try {
    return document.execCommand("copy") && copied;
  } finally {
    document.removeEventListener("copy", onCopy);
  }
}

function copyTextWithSelection(text) {
  const textarea = document.createElement("textarea");
  textarea.value = text;
  textarea.setAttribute("readonly", "");
  textarea.style.position = "fixed";
  textarea.style.top = "0";
  textarea.style.left = "0";
  textarea.style.width = "1px";
  textarea.style.height = "1px";
  textarea.style.opacity = "0";
  textarea.style.pointerEvents = "none";
  document.body.append(textarea);
  textarea.focus();
  textarea.select();
  textarea.setSelectionRange(0, textarea.value.length);
  const copied = document.execCommand("copy");
  textarea.remove();
  return copied;
}
