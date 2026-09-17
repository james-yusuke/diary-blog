(() => {
  const root = document.documentElement;
  const themeToggle = document.querySelector("[data-theme-toggle]");
  const themeStorageKey = "diary-theme";
  const themeModes = ["system", "light", "dark"];
  const themeLabels = { system: "自動", light: "ライト", dark: "ダーク" };
  const themeIcons = { system: "◐", light: "☀", dark: "◑" };
  const colorScheme = window.matchMedia("(prefers-color-scheme: dark)");
  let themeMode = "system";

  try {
    const storedTheme = window.localStorage.getItem(themeStorageKey);
    if (themeModes.includes(storedTheme)) themeMode = storedTheme;
  } catch {
    // A blocked storage area should not prevent the site from following the OS.
  }

  const effectiveTheme = () => themeMode === "system" ? (colorScheme.matches ? "dark" : "light") : themeMode;
  const nextThemeMode = () => {
    if (themeMode === "system") return effectiveTheme() === "dark" ? "light" : "dark";
    if (themeMode === "light") return "dark";
    return "system";
  };
  const renderTheme = () => {
    if (themeMode === "system") {
      delete root.dataset.theme;
    } else {
      root.dataset.theme = themeMode;
    }
    if (!themeToggle) return;

    const nextTheme = nextThemeMode();
    themeToggle.dataset.themeMode = themeMode;
    themeToggle.querySelector("[data-theme-icon]").textContent = themeIcons[themeMode];
    themeToggle.querySelector("[data-theme-label]").textContent = themeLabels[themeMode];
    themeToggle.setAttribute("aria-pressed", String(effectiveTheme() === "dark"));
    themeToggle.setAttribute("aria-label", `現在の表示テーマ: ${themeLabels[themeMode]}。${themeLabels[nextTheme]}に切り替える`);
    themeToggle.title = `表示テーマ: ${themeLabels[themeMode]}（クリックで${themeLabels[nextTheme]}）`;
  };

  renderTheme();
  themeToggle?.addEventListener("click", () => {
    themeMode = nextThemeMode();
    try {
      if (themeMode === "system") {
        window.localStorage.removeItem(themeStorageKey);
      } else {
        window.localStorage.setItem(themeStorageKey, themeMode);
      }
    } catch {
      // The selected theme still applies for this page when storage is unavailable.
    }
    renderTheme();
  });
  colorScheme.addEventListener("change", () => {
    if (themeMode === "system") renderTheme();
  });

  const dialog = document.querySelector("#command-palette");
  const trigger = document.querySelector("[data-command-trigger]");
  const input = document.querySelector("#command-query");
  const items = Array.from(document.querySelectorAll("[data-command-item]"));
  const close = document.querySelector("[data-command-close]");
  if (!dialog || !trigger || !input || !close) return;

  let activeIndex = 0;
  const visibleItems = () => items.filter((item) => !item.hidden);
  const setActive = (index) => {
    const visible = visibleItems();
    if (visible.length === 0) return;
    activeIndex = (index + visible.length) % visible.length;
    visible.forEach((item, itemIndex) => item.setAttribute("aria-selected", String(itemIndex === activeIndex)));
  };
  const filter = () => {
    const query = input.value.trim().toLocaleLowerCase();
    items.forEach((item) => {
      item.hidden = !item.textContent.toLocaleLowerCase().includes(query);
    });
    activeIndex = 0;
    setActive(0);
  };
  const open = () => {
    if (!dialog.open) dialog.showModal();
    input.value = "";
    filter();
    input.focus({ preventScroll: true });
  };
  const closePalette = () => {
    if (!dialog.open) return;
    dialog.close();
    trigger.focus({ preventScroll: true });
  };

  trigger.addEventListener("click", open);
  close.addEventListener("click", closePalette);
  input.addEventListener("input", filter);
  input.addEventListener("keydown", (event) => {
    if (event.key === "ArrowDown") {
      event.preventDefault();
      setActive(activeIndex + 1);
    }
    if (event.key === "ArrowUp") {
      event.preventDefault();
      setActive(activeIndex - 1);
    }
    if (event.key === "Enter" && input.value.trim() === "") {
      const item = visibleItems()[activeIndex];
      if (item) item.click();
    }
  });
  document.addEventListener("keydown", (event) => {
    const target = event.target;
    const isTyping = target instanceof HTMLInputElement || target instanceof HTMLTextAreaElement || target instanceof HTMLSelectElement;
    if ((event.metaKey || event.ctrlKey) && event.key.toLocaleLowerCase() === "k") {
      event.preventDefault();
      open();
    }
    if (event.key === "/" && !isTyping && !dialog.open) {
      event.preventDefault();
      open();
    }
    if (event.key === "Escape" && dialog.open) {
      event.preventDefault();
      closePalette();
    }
  });
})();
