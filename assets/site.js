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

(() => {
  const formats = {
    banner: [
      { key: '1a3cd0c7d36af5078c29bd8f41bdc95b', width: 728, height: 90 },
      { key: 'da18ef204e7b72146c8ce85420eac594', width: 468, height: 60 },
      { key: '065a5fd44299f5ebee83fa46c8f5db7d', width: 320, height: 50 },
    ],
    rectangle: [{ key: 'ea4a4dbf301abc0aee8bda933db1f8d5', width: 300, height: 250 }],
    sidebar: [{ key: '2366343916f8323dacb9bb48aaadb93f', width: 160, height: 300 }],
    tall: [{ key: '74be395aaed36f9088c2eb979381baea', width: 160, height: 600 }],
  };

  // Each banner gets its own document so atOptions and document.write cannot
  // overwrite another slot or the blog. Load only visible, nearby placements.
  document.querySelectorAll('[data-ad-slot]').forEach((slot) => {
    const mount = slot.querySelector('[data-ad-mount]');
    let selected;
    let nearby = false;
    const render = () => {
      const width = slot.getBoundingClientRect().width;
      const format = formats[slot.dataset.adSlot]?.find((item) => item.width <= width);
      if (!slot.getClientRects().length || !format) {
        mount.replaceChildren();
        mount.style.height = "0px";
        selected = undefined;
        slot.classList.remove('ad-slot--ready');
        return;
      }
      slot.classList.add('ad-slot--ready');
      mount.style.height = `${format.height}px`;
      if (!nearby || selected === format.key) return;
      selected = format.key;
      const frame = document.createElement('iframe');
      frame.title = '広告';
      frame.width = format.width;
      frame.height = format.height;
      frame.setAttribute('sandbox', 'allow-scripts allow-popups allow-popups-to-escape-sandbox');
      frame.setAttribute('scrolling', 'no');
      frame.srcdoc = `<!doctype html><html><head><meta name="viewport" content="width=device-width, initial-scale=1"><style>html,body{margin:0;padding:0;overflow:hidden}</style></head><body><script>var atOptions=${JSON.stringify({ ...format, format: 'iframe', params: {} })};<\/script><script src="https://www.highrevenueformat.com/${format.key}/invoke.js"><\/script></body></html>`;
      mount.replaceChildren(frame);
    };
    new ResizeObserver(render).observe(slot);
    const observer = new IntersectionObserver((entries) => {
      if (entries.some((entry) => entry.isIntersecting)) {
        nearby = true;
        render();
        observer.disconnect();
      }
    }, { rootMargin: '200px' });
    observer.observe(slot);
    render();
  });
})();
